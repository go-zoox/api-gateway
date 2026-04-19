# JSON 审计插件

包路径：`github.com/go-zoox/api-gateway/plugin/jsonaudit`

当 YAML 中 **`json_audit.enable`** 为 **`true`** 时会注册该插件。它在转发上游前**按需缓冲客户端请求体**（有上限），上游返回后在 **`OnResponse`** 中判断响应是否「**类似 JSON**」。只有满足条件时才会输出 **一条结构化 JSON 日志**，包含请求与响应内容（脱敏后），便于合规与安全审计。

## 能力概览

| 阶段 | 行为 |
| --- | --- |
| **`OnRequest`** | 路径与抽样通过后，读取请求体至多 **`max_body_bytes`**，并还原 **`Body`** 以便继续转发；把元数据与 body 保存在 **`ctx.Request.Context()`**。 |
| **`OnResponse`** | 读取上游响应体（同样上限）并还原 **`Body`**；若判定响应为 JSON 类，则序列化审计记录并通过 **`ctx.Logger.Infof`** 输出。 |

**注意：** 仅在**上游响应被判定为 JSON 类**时才会写审计日志；非此类响应**不会**为该请求生成审计条目。

## 何谓「响应为 JSON 类」？

满足以下**任一**条件：

1. **`Content-Type`** 中含子串 **`json`**（涵盖 `application/json`、`application/problem+json`、`application/vnd.api+json` 等）；**或**
2. **`sniff_json`** 为 **`true`**（默认），且去掉首尾空白后的 body 满足 **`json.Valid`**。

空响应体不会被判定为 JSON 类。

## gzip

当 **`decompress_gzip`** 为 **`true`**（默认），且 **`Content-Encoding`** 含 **`gzip`** 时，插件会尝试对响应体做一次解压（仍在 **`max_body_bytes`** 内），用于判定与记录。解压失败则退回原始字节继续处理（可能导致无法通过 JSON 校验）。

## 路径过滤

顺序如下：

1. **`include_paths` 非空**：请求路径必须以列表中**至少一条**前缀开头（前缀匹配）。
2. **`exclude_paths`**：若以其中**任意**前缀开头，则**不进行**审计。

可用于排除健康检查、二进制下载、流式或大体积接口。

## 抽样

**`sample_rate`** 表示在路径规则通过之后，有多少比例的请求参与审计：

- **`1`** 及以上 —— 尽可能记录（默认）。
- **`0`** 与 **`1`** 之间 —— 随机抽样（如 **`0.25`** ≈ 25%）。
- **`≤0`** —— 内部按 **`1`** 处理，即不限抽样（路径仍生效）。

未被抽中的请求不会写入审计上下文。

## 脱敏

若 body 可被解析为 JSON，则对对象键名做匹配（**不区分大小写**，**任意层级**），命中则将值替换为 **`"[REDACTED]"`**。

**`redact_keys` 为空**时使用内置默认值，包含：

`password`、`passwd`、`secret`、`token`、`authorization`、`api_key`、`apikey`、`access_token`、`refresh_token`。

若正文无法解析为 JSON，仍写入 **`request.body`** / **`response.body`**，一般为**字符串**片段；敏感 **HTTP 头**（如 **`Authorization`**、**`Cookie`**）及与 **`redact_keys`** 同名的 **query** 参数会被掩码。

## 日志字段说明

每条审计日志为 **一行 JSON**，由 **`ctx.Logger.Infof`** 输出（**info**）。

**顶层字段**

| 字段 | 含义 |
| --- | --- |
| **`type`** | 固定 **`json_audit`**。 |
| **`time`** | UTC 时间的字符串（`RFC3339Nano`）。 |
| **`timestamp`** | 与 **`time`** 同一时刻的 **Unix 毫秒时间戳**（整数），便于排序与数值检索。 |
| **`method`**、**`path`** | 与 **`request.method`** / **`request.path`** 相同（便于检索）。 |
| **`remote_addr`** | 客户端 `RemoteAddr`。 |
| **`request_id`** | 依次取第一个非空：`X-Request-ID`、`X-Correlation-ID`、`X-Trace-ID`。 |
| **`user_agent`** | **`User-Agent`** 请求头。 |
| **`response_status`** | 与 **`response.status`** 相同（上游状态码）。 |
| **`content_type`** | 上游响应 **`Content-Type`**。 |
| **`request_truncated`**、**`response_truncated`** | 是否因 **`max_body_bytes`** 截断 body。 |
| **`request`** | 见下表。 |
| **`response`** | 见下表。 |

**`request` 对象**

| 字段 | 含义 |
| --- | --- |
| **`method`**、**`path`** | HTTP 方法与路由路径（`ctx.Path`）。 |
| **`headers`** | 请求头，`map[string][]string`；内置敏感头（如 **`Authorization`**）值为 **`["[REDACTED]"]`**。 |
| **`query`** | URL 查询参数，`map[string][]string`；键名命中 **`redact_keys`** 时掩码。 |
| **`params`** | 路由参数 **`ctx.Params().ToMap()`**，无则为空对象 **`{}`**。 |
| **`body`** | 请求体：可为解析后的 JSON（键脱敏），或解析失败时的字符串。 |

**`response` 对象**

| 字段 | 含义 |
| --- | --- |
| **`status`** | 上游 HTTP 状态码。 |
| **`body`** | 响应体：解析后的 JSON（键脱敏），或字符串。 |

请结合现有日志采集与合规策略使用；涉密环境勿依赖默认脱敏作为唯一防护。

## 示例：配置与日志

### 场景说明

客户端 **`POST /api/v1/login`**，请求体为 JSON（含账号口令）；上游返回 **`200`**，响应体为 JSON（含会话 **`token`**）。网关开启 **`json_audit`**，并使用默认脱敏键（含 **`password`**、**`token`** 等）。

只有在本请求通过**路径规则**与**抽样**之后才会记录（此处假设通过）。插件输出 **一条 info 级别日志**，其消息内容为**单行 JSON 对象**（实现上为 `ctx.Logger.Infof("%s", …)`）。

实际环境里，日志前面可能还会带 **级别、时间、logger 名称** 等前缀（取决于 Zoox 与采集配置）；下面示例仅展示 **审计对象的 JSON 正文**。

### 网关配置片段

最小启用：

```yaml
port: 8080

json_audit:
  enable: true
```

更接近生产的写法（缩小路径范围 + 自定义脱敏字段）：

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

# … routes、backend、cache 等与 json_audit 无直接耦合 …
```

若配置了 **`include_paths: [/api/v1/]`**，则 **`/api/v1/login`** 可走审计（未被 **`exclude_paths`** 排除时）；**`/health`** 等被排除的路径不会产生审计日志。

### 示例审计日志（正文 JSON）

下文为便于阅读做了**换行与缩进**；线上一般为**单行紧凑**输出。

**请求体含义：** `{"username":"alice","password":"secret123"}`  
**响应体含义：** `{"ok":true,"token":"eyJhbG..."}`  

**写入日志的审计对象**（`password`、`token` 已在 body 中脱敏；**`Authorization`** 在 headers 中脱敏）：

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

客户端若携带 **`X-Request-ID`**、**`X-Correlation-ID`** 或 **`X-Trace-ID`**，第一个非空值写入 **`request_id`**（若请求里仍有对应头，也会出现在 **`request.headers`**）。

若请求或响应体超过 **`max_body_bytes`**，会被截断，**`request_truncated`** / **`response_truncated`** 为 **`true`**。

当响应仍被视为 JSON 类但正文无法按 JSON 解析时，**`request.body`** / **`response.body`** 可能为**字符串**而非对象。

## 更多配置片段

最小启用：

```yaml
json_audit:
  enable: true
```

路径 + 抽样 + 脱敏（与上文「生产常用」一致）：

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

### 字段说明（`json_audit`）

**`json_audit`** 对应根配置里的 **`config.JSONAudit`**（见 [配置 API](/zh/api/config)）。插件在 **`Prepare`** 中直接使用 **`cfg.JSONAudit`**，不再在插件包内重复定义配置类型。

YAML 使用 **snake_case**。

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
<th align="left">字段</th>
<th align="left">是否必须</th>
<th align="left">默认值</th>
<th align="left">简要说明</th>
</tr>
</thead>
<tbody>
<tr>
<td valign="top"><code>enable</code></td>
<td valign="top">是*</td>
<td valign="top"><code>false</code></td>
<td valign="top">置为 <code>true</code> 才会加载插件。</td>
</tr>
<tr>
<td valign="top"><code>max_body_bytes</code></td>
<td valign="top">否</td>
<td valign="top"><code>1048576</code></td>
<td valign="top">请求/响应体采集的最大字节数。</td>
</tr>
<tr>
<td valign="top"><code>sample_rate</code></td>
<td valign="top">否</td>
<td valign="top"><code>1</code></td>
<td valign="top">路径过滤后的抽样比例；≤0 等价于全量。</td>
</tr>
<tr>
<td valign="top"><code>sniff_json</code></td>
<td valign="top">否</td>
<td valign="top"><code>true</code></td>
<td valign="top">无 JSON 的 Content-Type 时是否用 <code>json.Valid</code> 嗅探。</td>
</tr>
<tr>
<td valign="top"><code>decompress_gzip</code></td>
<td valign="top">否</td>
<td valign="top"><code>true</code></td>
<td valign="top"><code>Content-Encoding</code> 含 gzip 时是否尝试解压。</td>
</tr>
<tr>
<td valign="top"><code>include_paths</code></td>
<td valign="top">否</td>
<td valign="top"><em>空</em></td>
<td valign="top">前缀白名单；空表示在排除前不过滤路径。</td>
</tr>
<tr>
<td valign="top"><code>exclude_paths</code></td>
<td valign="top">否</td>
<td valign="top"><em>空</em></td>
<td valign="top">前缀黑名单，在 include 之后评估。</td>
</tr>
<tr>
<td valign="top"><code>redact_keys</code></td>
<td valign="top">否</td>
<td valign="top"><em>内置列表</em></td>
<td valign="top">JSON 对象键（任意深度）命中则掩码。</td>
</tr>
</tbody>
</table>
</div>

\*仅当 **`enable: true`** 时注册插件。

## 限制与注意

- **全量缓冲**：请求与响应会在内存中缓冲至 **`max_body_bytes`**；流式、超大 body 会占用内存或被截断。
- **`Content-Type` 声明为 JSON 但正文非法**：仍可能按「JSON 类」处理，**`request.body`** / **`response.body`** 可能为**字符串**，且无键级脱敏。
- **抽样**为随机过程，短时间窗口内比例会有波动。

## 相关链接

- [插件总览](./)
- [配置说明](/zh/guide/configuration)
