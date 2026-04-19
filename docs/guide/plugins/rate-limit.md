# Rate limiting plugin

Package: `github.com/go-zoox/api-gateway/plugin/ratelimit`

The rate limiting plugin is registered when **either** the global `rate_limit.enable` flag is true **or** at least one route sets `rate_limit.enable`. It runs in `OnRequest` before traffic reaches backends and can return **429 Too Many Requests** when a client exceeds its quota.

## Features

- **Keys**: IP (with `X-Forwarded-For` / `X-Real-IP` support), user id (Bearer / `X-User-ID`), API key (`X-API-Key`, `Authorization: ApiKey …`, or `api_key` query), or a custom header.
- **Algorithms**: `token-bucket`, `leaky-bucket`, `fixed-window`.
- **Counters**: Stored only via **`zoox.Application.Cache()`** (set top-level `cache` in YAML for Redis; otherwise the framework’s in-memory KV).
- **Scope**: Global defaults plus **per-route** overrides.

## Configuration

### Global

```yaml
cache:
  host: redis.example.com
  port: 6379
# ...
rate_limit:
  enable: true
  algorithm: token-bucket
  key_type: ip
  limit: 100
  window: 60
  burst: 20
  message: "Rate limit exceeded"
```

### Per route

Route-level `rate_limit` overrides the global policy for matching paths.

```yaml
routes:
  - name: user-service
    path: /v1/user
    rate_limit:
      enable: true
      algorithm: token-bucket
      key_type: user
      limit: 10
      window: 60
      burst: 5
    backend:
      service:
        name: user-service
```

### Field reference (`rate_limit`)

YAML keys use **snake_case** (e.g. `key_type`). **Required?** is whether you must set the field for an effective rate-limit policy; **Default** is the value when the field is omitted (from struct tags / zero values). The **Summary** column is short; each field has a dedicated section under [Field details](#field-details).

<div style="overflow-x:auto">
<table style="table-layout:fixed;width:100%;max-width:52rem;border-collapse:collapse">
<colgroup>
<col style="width:7.5rem" />
<col style="width:5rem" />
<col style="width:6.5rem" />
<col style="width:22rem" />
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
<td valign="top"><code>limit</code></td>
<td valign="top">Yes</td>
<td valign="top">—</td>
<td valign="top">Max requests counted per active window for this policy. <a href="#field-limit">Details</a></td>
</tr>
<tr>
<td valign="top"><code>window</code></td>
<td valign="top">Yes</td>
<td valign="top">—</td>
<td valign="top">Window length in seconds; drives refill/leak behaviour with <code>limit</code>. <a href="#field-window">Details</a></td>
</tr>
<tr>
<td valign="top"><code>enable</code></td>
<td valign="top">No</td>
<td valign="top"><code>false</code></td>
<td valign="top">Turns the plugin on for global and/or route scope. <a href="#field-enable">Details</a></td>
</tr>
<tr>
<td valign="top"><code>algorithm</code></td>
<td valign="top">No</td>
<td valign="top"><code>token-bucket</code></td>
<td valign="top">Which limiter implementation runs (<code>token-bucket</code>, <code>leaky-bucket</code>, <code>fixed-window</code>). <a href="#field-algorithm">Details</a></td>
</tr>
<tr>
<td valign="top"><code>key_type</code></td>
<td valign="top">No</td>
<td valign="top"><code>ip</code></td>
<td valign="top">How the per-client rate-limit key is derived from the request. <a href="#field-key-type">Details</a></td>
</tr>
<tr>
<td valign="top"><code>key_header</code></td>
<td valign="top">No</td>
<td valign="top"><em>(empty)</em></td>
<td valign="top">Header name when <code>key_type</code> is <code>header</code>. <a href="#field-key-header">Details</a></td>
</tr>
<tr>
<td valign="top"><code>burst</code></td>
<td valign="top">No</td>
<td valign="top"><code>0</code></td>
<td valign="top">Token-bucket bucket capacity; other algorithms usually ignore it. <a href="#field-burst">Details</a></td>
</tr>
<tr>
<td valign="top"><code>message</code></td>
<td valign="top">No</td>
<td valign="top"><code>Too Many Requests</code></td>
<td valign="top">Plain-text body returned with HTTP 429 when blocked. <a href="#field-message">Details</a></td>
</tr>
<tr>
<td valign="top"><code>headers</code></td>
<td valign="top">No</td>
<td valign="top"><em>(empty map)</em></td>
<td valign="top">Extra HTTP headers attached only to 429 responses. <a href="#field-headers">Details</a></td>
</tr>
</tbody>
</table>
</div>

**Where to put fields:** Use the top-level `rate_limit:` block for defaults; add `rate_limit:` under a route to override for paths that match that route (see [Route matching precedence](#route-matching-precedence)).

## Field details

<a id="field-details"></a>

Each subsection describes one YAML field: meaning, default, usage, and an example snippet.

<a id="field-limit"></a>
### `limit`

- **Meaning:** Maximum number of requests **allowed** within one policy window for a single rate-limit **key** (see `key_type`). The exact semantics depend on `algorithm`, but it always acts as the primary quota number (e.g. sustained rate or fixed-window cap).
- **Default:** No default — the field is **required** for an effective policy. If set to zero or negative, the policy is skipped (fail-open).
- **Usage:** Set higher limits on trusted routes or admin APIs; tighten per-route overrides for expensive endpoints. Pair with `window` to express “N requests per window seconds”.
- **Example:** Allow at most **100** counted requests per active window (with `window` defining the window length):

```yaml
rate_limit:
  enable: true
  limit: 100
  window: 60
```

<a id="field-window"></a>
### `window`

- **Meaning:** Length of the policy time window in **seconds** (integer). With **token-bucket**, refill rate is `limit / window`. With **fixed-window**, counts reset when the stored window expires. With **leaky-bucket**, leak rate uses the same ratio.
- **Default:** No default — **required**. If zero or negative, the policy is skipped.
- **Usage:** Short windows react quickly to bursts; long windows smooth traffic over minutes or hours. Must be consistent with how clients retry (see `Retry-After`).
- **Example:** A **60**-second window with `limit: 100` caps at 100 requests per minute (behaviour varies slightly by algorithm):

```yaml
rate_limit:
  enable: true
  limit: 100
  window: 60
```

<a id="field-enable"></a>
### `enable`

- **Meaning:** Enables the plugin for the **scope** where it appears: the root `rate_limit` block (global defaults) or a route’s `rate_limit` (override for matching paths).
- **Default:** `false`. The plugin is registered only if **at least one** `rate_limit.enable: true` exists (global or any route).
- **Usage:** Turn on globally and disable specific routes only if your schema supports it; or leave global off and enable only sensitive routes.
- **Example:** Enable only on one route (global `rate_limit` omitted or `enable: false`):

```yaml
routes:
  - path: /v1/expensive
    rate_limit:
      enable: true
      limit: 20
      window: 60
```

<a id="field-algorithm"></a>
### `algorithm`

- **Meaning:** Which limiter implementation runs. Values: `token-bucket` (default), `leaky-bucket`, `fixed-window` — see [Algorithms (summary)](#algorithms-summary). Unknown values fall back to **token-bucket** in the factory.
- **Default:** `token-bucket` when omitted.
- **Usage:** Use **fixed-window** for simple counting in cache; **token-bucket** for refill + burst; **leaky-bucket** for smoothing (rate-like behaviour).
- **Example:** Use a simple fixed window for an IP-keyed public API:

```yaml
rate_limit:
  enable: true
  algorithm: fixed-window
  key_type: ip
  limit: 50
  window: 60
```

<a id="field-key-type"></a>
### `key_type`

- **Meaning:** How the per-client rate-limit **key** is derived. Values: `ip`, `user`, `apikey`, `header`. Any other string is treated like **`ip`**.
- **Default:** `ip` when omitted.
- **Details:** **`ip`** — first `X-Forwarded-For` hop, then `X-Real-IP`, then `RemoteAddr`. **`user`** — `Authorization: Bearer` token value, then `X-User-ID`; else falls back like `ip`. **`apikey`** — `X-API-Key`, then `Authorization: ApiKey …`, then query `api_key`; else IP. **`header`** — uses `key_header`; if empty, falls back like `ip`.
- **Usage:** Choose `ip` for anonymous traffic; `user` or `apikey` for authenticated quotas; `header` for tenancy or custom routing headers.
- **Example:** Rate limit **per API key** (header `X-API-Key` wins first):

```yaml
rate_limit:
  enable: true
  key_type: apikey
  limit: 1000
  window: 3600
```

<a id="field-key-header"></a>
### `key_header`

- **Meaning:** HTTP header **name** (not value) when `key_type: header`. The rate-limit key includes that header’s value.
- **Default:** Empty string. With `key_type: header` and an empty name, extraction falls back like **`ip`**.
- **Usage:** Set to stable tenant or client identifiers (e.g. `X-Tenant-ID`). Avoid highly volatile headers unless intentional.
- **Example:** One quota bucket per **tenant id** carried in `X-Tenant-ID`:

```yaml
rate_limit:
  enable: true
  key_type: header
  key_header: X-Tenant-ID
  limit: 500
  window: 60
```

<a id="field-burst"></a>
### `burst`

- **Meaning:** For **token-bucket**, maximum **bucket capacity** (burst size). Refill rate stays `limit / window`. If `burst` ≤ 0 or omitted, capacity defaults to **`limit`**. **Leaky-bucket** / **fixed-window** may ignore this field.
- **Default:** `0` (meaning “use `limit` as bucket capacity” for token-bucket).
- **Usage:** Set `burst` **greater than** `limit` only when you want a larger short-term spike than sustained `limit`/`window` alone.
- **Example:** Sustain ~10 req/s (`limit: 10`, `window: 1`) but allow up to **50** concurrent burst tokens:

```yaml
rate_limit:
  enable: true
  algorithm: token-bucket
  limit: 10
  window: 1
  burst: 50
```

<a id="field-message"></a>
### `message`

- **Meaning:** Response **body** when the gateway returns **429 Too Many Requests** for rate limiting.
- **Default:** `Too Many Requests` when omitted or empty (depending on gateway handling of empty strings).
- **Usage:** Use plain text or a JSON string your clients parse consistently with other errors.
- **Example:** Return a small JSON payload (quote the whole value in YAML):

```yaml
rate_limit:
  enable: true
  limit: 60
  window: 60
  message: '{"error":"too_many_requests","retry":true}'
```

<a id="field-headers"></a>
### `headers`

- **Meaning:** Extra **response headers** sent **only on 429**, in addition to `X-RateLimit-*` / `Retry-After` when the writer is available.
- **Default:** Empty map — no extra headers.
- **Usage:** Add policy names, hints, or correlation ids — never secrets.
- **Example:** Attach a stable policy label for monitoring:

```yaml
rate_limit:
  enable: true
  limit: 100
  window: 60
  headers:
    X-Rate-Policy: standard-tier
```


## Route matching precedence

When multiple routes define rate limits:

1. Candidate routes are sorted by **path length (longest first)**.
2. The request path must **exactly match** a route path, or match as a **prefix** where the next character is `/` (so `/users` matches `/users/123` but not `/users-extra`).

Longer paths win over shorter prefixes (for example `/api/v1` wins over `/api` for `/api/v1/foo`).

## Algorithms (summary)

| Algorithm | Behaviour |
| --- | --- |
| `token-bucket` | Allows bursts up to `burst`; refills against `limit` / `window`. |
| `leaky-bucket` | Smooth throughput; burst is not used as extra capacity in the same way as token bucket. |
| `fixed-window` | Simple counting window aligned to storage implementation. |

## Cache / KV backend

Counters live in **`zoox.Application.Cache()`** — the same `cache.Cache` instance the framework builds from `Config.Cache` (`cache.New`). Gateway **`prepare()`** writes `Config.Cache` when your YAML sets **`cache`** (e.g. Redis). If you omit **`cache`**, zoox still exposes `Application.Cache()` using its default **in-memory** KV engine (not a separate plugin-owned map).

There is **no `rate_limit.storage`** knob and no storage-type branching in code: counters use **`newCacheStorage(app.Cache())`** only.

## Response headers

Successful passes and many responses include:

- `X-RateLimit-Limit`
- `X-RateLimit-Remaining`
- `X-RateLimit-Reset` (Unix timestamp)

When returning **429**, the gateway also sets `Retry-After` (seconds) when possible, plus any custom keys from `headers`.

If the algorithm or storage returns an error, the plugin **allows** the request (fail-open) and logs the error.

## Related

- [Plugins overview](./)
- [Configuration](/guide/configuration)
- [Routing](/guide/routing)
