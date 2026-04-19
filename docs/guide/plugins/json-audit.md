# JSON audit plugin

Package: `github.com/go-zoox/api-gateway/plugin/jsonaudit`

The gateway registers the JSON audit plugin when **top-level** **`json_audit.enable`** is **`true`** **or** **any route** sets **`json_audit.enable`** (same idea as rate limiting). It buffers the **incoming request body** (bounded) before the upstream call, then after the upstream responds checks whether the **response looks like JSON**. Only then it emits **one structured JSON log line** containing both request and response payloads (after redaction), suitable for compliance / security audits.

## Behaviour summary

| Phase | What happens |
| --- | --- |
| **`OnRequest`** | If path rules and sampling allow, read the client request body up to **`max_body_bytes`**, restore `Body` for forwarding, stash metadata + body on **`ctx.Request.Context()`**. |
| **`OnResponse`** | Read the upstream response body (same size cap), restore `Body` for the client. **If** the response is deemed JSON-like, marshal an audit record and log via **`ctx.Logger.Infof`**. |

**Important:** Audit lines are emitted **only when the upstream response qualifies as JSON-like** (see below). Non-JSON responses produce **no** audit record for that request.

## When is a response JSON-like?

Either:

1. **`Content-Type`** indicates JSON — contains the substring **`json`** (covers `application/json`, `application/problem+json`, `application/vnd.api+json`, etc.), **or**
2. **`sniff_json`** is **`true`** (default) **and** the trimmed body passes Go’s **`json.Valid`**.

Empty response bodies never qualify.

## gzip

If **`decompress_gzip`** is **`true`** (default) **and** `Content-Encoding` contains **`gzip`**, the plugin attempts one-shot gzip decompression (still bounded by **`max_body_bytes`**) **for detection and logging**. If decompression fails, the raw bytes are used instead (JSON detection may fail).

## Path filtering

Evaluation order:

1. If **`include_paths`** is non-empty, the request path must start with **at least one** listed prefix (prefix match).
2. **`exclude_paths`**: if the path starts with **any** listed prefix, auditing is skipped for that request.

Use **`exclude_paths`** for health checks, binaries, streaming endpoints, or routes with very large payloads.

## Sampling

**`sample_rate`** is the fraction of requests considered for auditing after path checks:

- **`1.0`** or greater — consider every eligible path (default).
- Between **`0`** and **`1`** — Bernoulli sample (e.g. **`0.1`** ≈ 10%).
- **`0`** or negative — treated like **`1.0`** internally (audit all qualifying paths).

Skipped requests carry no audit payload for that hop.

## Redaction

Bodies are parsed as JSON when possible; object keys matched (case-insensitive, any nesting depth) are replaced by **`"[REDACTED]"`** in the logged structures.

If **`redact_keys`** is **empty**, the plugin uses built-in defaults, including:

`password`, `passwd`, `secret`, `token`, `authorization`, `api_key`, `apikey`, `access_token`, `refresh_token`.

Non-JSON bodies are still logged under **`request.body`** / **`response.body`**: valid JSON becomes a parsed tree (after key redaction); otherwise a **string** fragment is stored. Sensitive **HTTP headers** (`Authorization`, `Cookie`, …) and **query** keys that match **`redact_keys`** are masked.

## Audit log schema

Each audit line is one JSON object passed to **`ctx.Logger.Infof`** (level **info**).

**Top-level fields**

| Field | Meaning |
| --- | --- |
| **`type`** | Always **`json_audit`**. |
| **`time`** | UTC wall time as string (`RFC3339Nano`). |
| **`timestamp`** | Same instant as Unix **milliseconds** (`int`), for sorting / numeric pipelines. |
| **`method`**, **`path`** | Same as **`request.method`** / **`request.path`** (shortcut for indexing). |
| **`remote_addr`** | Client `RemoteAddr`. |
| **`request_id`** | First non-empty among **`X-Request-ID`**, **`X-Correlation-ID`**, **`X-Trace-ID`**. |
| **`user_agent`** | `User-Agent` header. |
| **`response_status`** | Same as **`response.status`** (shortcut). |
| **`content_type`** | Upstream response **`Content-Type`**. |
| **`request_truncated`**, **`response_truncated`** | Whether body capture hit **`max_body_bytes`**. |
| **`request`** | See below. |
| **`response`** | See below. |

**`request` object**

| Field | Meaning |
| --- | --- |
| **`method`**, **`path`** | HTTP method and routed path (`ctx.Path`). |
| **`headers`** | Request headers as **`map[string][]string`**; known sensitive headers replaced with **`["[REDACTED]"]`**. |
| **`query`** | URL query (`map[string][]string`); parameter names matching **`redact_keys`** are redacted. |
| **`params`** | Route parameters from **`ctx.Params().ToMap()`** (`map[string]any`), empty object if none. |
| **`body`** | Request body: parsed JSON (after redaction) when valid, else a raw string. |

**`response` object**

| Field | Meaning |
| --- | --- |
| **`status`** | HTTP status code from upstream. |
| **`body`** | Response body: parsed JSON after redaction when valid, else a raw string. |

Collect logs with your existing pipeline (stdout, shipper, SIEM). Avoid logging truly secret environments in plaintext.

## Example: configuration and log output

### Scenario

A client calls **`POST /api/v1/login`** with a JSON body that contains credentials. The upstream returns **`200`** and a JSON body that includes a session token. The gateway has **`json_audit`** enabled with default redaction keys (`password`, `token`, …).

Auditing runs only if this request passes **path filters** and **sampling** (here we assume it does). The plugin writes **one info-level log message** whose message payload is a **single JSON object** (see implementation: `ctx.Logger.Infof("%s", …)`).

Your log collector may still prefix each line with **severity**, **timestamp**, or **logger name** depending on Zoox / deployment settings—the example below shows **only the audit JSON payload**.

### Sample gateway YAML (excerpt)

Minimal toggle:

```yaml
port: 8080

json_audit:
  enable: true
```

Typical production-style snippet (narrow paths + custom redaction):

```yaml
port: 8080

json_audit:
  enable: true
  max_body_bytes: 1048576
  sample_rate: 1
  sniff_json: true
  decompress_gzip: true
  include_paths:
    - /api/v1/
  exclude_paths:
    - /health
    - /metrics
  redact_keys:
    - password
    - secret
    - national_id

# … routes / backend / cache etc. — unchanged by json_audit …
```

With **`include_paths: [/api/v1/]`**, a request to **`/api/v1/login`** is audited (unless excluded); **`GET /health`** is not.

### Sample audit log payload

Below is **pretty-printed for readability**. In production it is usually emitted as **one compact line**.

**Client request body (conceptual):** `{"username":"alice","password":"secret123"}`  
**Upstream response body (conceptual):** `{"ok":true,"token":"eyJhbG..."}`  

**Recorded audit object** (password / token redacted in bodies; **`Authorization`** redacted in headers):

```json
{
  "type": "json_audit",
  "time": "2026-04-19T14:32:01.234567891Z",
  "timestamp": 1776609121234,
  "method": "POST",
  "path": "/api/v1/login",
  "remote_addr": "203.0.113.50:49152",
  "request_id": "req-7f91ac",
  "user_agent": "ExampleClient/1.0",
  "response_status": 200,
  "content_type": "application/json; charset=utf-8",
  "request_truncated": false,
  "response_truncated": false,
  "request": {
    "method": "POST",
    "path": "/api/v1/login",
    "headers": {
      "Accept": ["application/json"],
      "Authorization": ["[REDACTED]"],
      "Content-Type": ["application/json"],
      "User-Agent": ["ExampleClient/1.0"],
      "X-Request-ID": ["req-7f91ac"]
    },
    "query": {
      "source": ["web"]
    },
    "params": {},
    "body": {
      "username": "alice",
      "password": "[REDACTED]"
    }
  },
  "response": {
    "status": 200,
    "body": {
      "ok": true,
      "token": "[REDACTED]"
    }
  }
}
```

If the client sends **`X-Request-ID`**, **`X-Correlation-ID`**, or **`X-Trace-ID`**, the first non-empty value appears in **`request_id`** (and is still listed under **`request.headers`** when present).

If either body exceeds **`max_body_bytes`**, the captured bytes are truncated and **`request_truncated`** or **`response_truncated`** is **`true`**.

When JSON parsing fails but the response is still treated as JSON-like (for example **`Content-Type: application/json`** with invalid bytes), **`request.body`** / **`response.body`** may be a **string** instead of an object.

## Configuration snippets

Minimal:

```yaml
json_audit:
  enable: true
```

Stricter routing and masking (same as in the walkthrough):

```yaml
json_audit:
  enable: true
  max_body_bytes: 524288
  sample_rate: 0.25
  sniff_json: true
  decompress_gzip: true
  include_paths:
    - /api/v1/
  exclude_paths:
    - /health
    - /metrics
  redact_keys:
    - password
    - secret
    - national_id
```

### Field reference (`json_audit`)

YAML **`json_audit`** maps to **`config.JSONAudit`** / **`route.JSONAudit`** (see [Config API](/api/config)). Use the **root** block for defaults and optional **`routes[].json_audit`** for overrides; the plugin resolves **longest-prefix** route settings per request, then falls back to the global block when **`enable`** is true — one struct type, no separate **`jsonaudit.Config`** file.

YAML keys use **snake_case**.

<div style="overflow-x:auto">
<table style="table-layout:fixed;width:100%;max-width:52rem;border-collapse:collapse">
<colgroup>
<col style="width:8rem" />
<col style="width:5rem" />
<col style="width:7rem" />
<col style="width:20rem" />
</colgroup>
<thead>
<tr>
<th align="left">Field</th>
<th align="left">Required?</th>
<th align="left">Default</th>
<th align="left">Summary</th>
</tr>
</thead>
<tbody>
<tr>
<td valign="top"><code>enable</code></td>
<td valign="top">Yes*</td>
<td valign="top"><code>false</code></td>
<td valign="top">Must be <code>true</code> for the plugin to register.</td>
</tr>
<tr>
<td valign="top"><code>max_body_bytes</code></td>
<td valign="top">No</td>
<td valign="top"><code>1048576</code></td>
<td valign="top">Max bytes read from request/response bodies for auditing.</td>
</tr>
<tr>
<td valign="top"><code>sample_rate</code></td>
<td valign="top">No</td>
<td valign="top"><code>1</code></td>
<td valign="top">Fraction of requests to audit after path filtering; ≤0 behaves like full sampling.</td>
</tr>
<tr>
<td valign="top"><code>sniff_json</code></td>
<td valign="top">No</td>
<td valign="top"><code>true</code></td>
<td valign="top">Allow <code>json.Valid</code> sniff when Content-Type is not JSON.</td>
</tr>
<tr>
<td valign="top"><code>decompress_gzip</code></td>
<td valign="top">No</td>
<td valign="top"><code>true</code></td>
<td valign="top">Try gzip decompress when <code>Content-Encoding</code> indicates gzip.</td>
</tr>
<tr>
<td valign="top"><code>include_paths</code></td>
<td valign="top">No</td>
<td valign="top"><em>(empty)</em></td>
<td valign="top">Prefix allow list; empty means all paths (minus excludes).</td>
</tr>
<tr>
<td valign="top"><code>exclude_paths</code></td>
<td valign="top">No</td>
<td valign="top"><em>(empty)</em></td>
<td valign="top">Prefix deny list evaluated after includes.</td>
</tr>
<tr>
<td valign="top"><code>redact_keys</code></td>
<td valign="top">No</td>
<td valign="top"><em>(built-ins)</em></td>
<td valign="top">JSON object keys (any depth) replaced with <code>[REDACTED]</code>.</td>
</tr>
</tbody>
</table>
</div>

\*The plugin registers only when **`enable: true`**.

## Limitations

- **Buffered bodies:** Request and response must be buffered in memory up to **`max_body_bytes`**; streaming / very large payloads can raise memory usage or truncate.
- **`application/json` with invalid body:** Still counted as JSON-like via media type — **`response.body`** / **`request.body`** may be plain **strings** without structured key redaction.
- **Sampling** is probabilistic — not guaranteed uniform over short windows.

## Related

- [Plugins overview](./)
- [Configuration](/guide/configuration)
