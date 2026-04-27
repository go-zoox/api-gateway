# JSON 审计插件

包路径：`github.com/go-zoox/api-gateway/plugin/jsonaudit`

当 **顶层** **`json_audit.enable`** 为真，**或** **任一路由** 启用 **`json_audit.enable`** 时，会注册该插件（与限流的路由级启用方式一致）。它在转发上游前**按需缓冲客户端请求体**（有上限），上游返回后在 **`OnResponse`** 中判断响应是否「**类似 JSON**」。只有满足条件时才会输出 **一条结构化 JSON 日志**，包含请求与响应内容（是否脱敏取决于 `redact` 配置与 provider 默认策略），便于合规与安全审计。

## 能力概览

| 阶段 | 行为 |
| --- | --- |
| **`OnRequest`** | 路径与抽样通过后，读取请求体至多 **`max_body_bytes`**，并还原 **`Body`** 以便继续转发；把元数据与 body 保存在 **`ctx.Request.Context()`**。 |
| **`OnResponse`** | 读取上游响应体（同样上限）并还原 **`Body`**；若判定响应为 JSON 类，则序列化 **一行 JSON 审计记录**，并按 **`json_audit.output`**（**`output.provider`**，默认 **console** → 应用日志 **info**）写入对应渠道。 |

**注意：** 仅在**上游响应被判定为 JSON 类**时才会写审计日志；非此类响应**不会**为该请求生成审计条目。

## 配置说明（`json_audit`）

**`json_audit`** 对应 **`config.JSONAudit`** / **`route.JSONAudit`**（见 [配置 API](/zh/api/config)）。根配置块作默认值，可在 **`routes[].json_audit`** 覆盖；插件对每个请求按 **最长前缀匹配** 路由后回退到全局 **`enable`** 为真时的全局块 —— 单一结构体类型，插件包内不再重复定义。

YAML 使用 **snake_case**。

### 字段一览

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
<td valign="top"><code>output</code></td>
<td valign="top">否</td>
<td valign="top"><code>provider: console</code></td>
<td valign="top">嵌套块：<strong><code>provider</code></strong> — <code>console</code>（默认）、<code>file</code>、<code>http</code>、<code>database</code>（<code>webhook</code> / <code>endpoint</code> / <code>api</code> 视为 http；<code>db</code> / <code>sql</code> 视为 database）。<code>provider: file</code> 时配置 <strong><code>file.path</code></strong>；<code>provider: http</code> 时配置 <strong><code>http</code></strong>（<code>url</code> 必填；可选 <code>method</code>、<code>headers</code>、<code>timeout_seconds</code>）；<code>provider: database</code> 时必须配置独立的 <strong><code>database</code></strong>（支持 DSN 或结构化字段），不回退顶层 <code>database</code>。使用 URL 形式 DSN（如 <code>postgres://</code>、<code>mysql://</code>）时可省略 <code>engine</code>；结构化字段或非 URL DSN 时需显式指定 <code>engine</code>。若选择 <code>database</code> 但未配置 <code>output.database</code>，启动阶段会 panic。数据库模式通过 <code>github.com/go-zoox/gormx</code> 建连，并在启动时自动执行迁移。HTTP/数据库写入失败会退回 **info** 控制台输出。</td>
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
<td valign="top"><code>redact</code></td>
<td valign="top">否</td>
<td valign="top"><code>enable</code> 默认按 provider*</td>
<td valign="top">嵌套块：<strong><code>enable</code></strong> 控制是否脱敏（显式配置优先）；未显式配置时默认仅 <code>provider=console</code> 开启，<code>file/http/database</code> 默认关闭。<strong><code>keys</code></strong> 为 JSON/query 键名列表，空则用内置列表。<code>enable: false</code> 时头、query、JSON 正文均<strong>不脱敏</strong>（含 <code>Authorization</code> 明文）。</td>
</tr>
</tbody>
</table>
</div>

\*顶层 **`json_audit.enable`** 为真，或**任一路由**启用 **`json_audit`** 时注册插件。

\*\***`redact.enable`** 显式值优先；若省略，默认仅 **`output.provider=console`** 开启脱敏，其他 sink 默认关闭。

### 字段详解

以下示例默认写在**根配置**的 **`json_audit:`** 下；若写在 **`routes[].json_audit`**，则表示按**最长前缀**匹配到该路由时生效、可覆盖全局。字段名均为 **snake_case**。

#### `enable`

**`true`** 时参与启用插件：**顶层** **`json_audit.enable`** 或**任一路由** **`json_audit.enable`** 为真时都会注册插件（与限流一致）。在某一 **`routes`** 条目中设置 **`enable: true`**，表示该路由前缀下可使用该块里的审计配置；每个请求实际生效的配置按**最长前缀匹配**路由，再回退到根配置。

```yaml
json_audit:
  enable: true
```

#### `output`（嵌套）

由 **`output.provider`** 决定**每条审计记录**（单行 JSON）写到哪里：

| `provider` | 行为 |
| --- | --- |
| **`console`**（默认） | 走应用 **`info`** 日志（与其它网关日志同一套管线）。 |
| **`file`** | 追加写入 **`output.file.path`**，每条记录后换行（NDJSON）。 |
| **`http`** | 向 **`output.http.url`** 发起请求（默认 **`POST`**），请求体为审计 JSON，**`Content-Type: application/json`**。 |
| **`database`** | 使用 **`output.database`**（支持 **DSN** 或**完整字段配置**）通过 `gormx` 将审计记录写入自动迁移的数据库表。 |

**`webhook`**、**`endpoint`**、**`api`** 写在 **`provider`** 上时与 **`http`** 等价；**`db`**、**`sql`** 与 **`database`** 等价。若 **`provider`** 为 **`file`** / **`http`** / **`database`**，启动时会校验对应必填项：**`file.path`**、**`http.url`**、或数据库连接字段。数据库模式支持两种写法：**`output.database.dsn`** 或结构化字段（**`engine`**、**`host`**、**`port`**、**`username`**、**`password`**、**`db`**）；结构化字段优先级高于 `dsn`。当 `dsn` 是 URL 形式（如 `postgres://...`、`mysql://...`、`sqlite:///...`）时可从 scheme 自动识别引擎，`engine` 可省略；若为非 URL DSN（如 SQLite 文件路径、MySQL 旧 DSN）则通常需要显式 `engine`。数据库配置仅来自 **`output.database`**，不再回退读取顶层 **`database`**。若 **`provider: database`** 但 **`output.database`** 完全缺失，插件会在启动阶段直接 panic。数据库模式会在启动时通过 `gormx` 建立连接并执行 `AutoMigrate`，并按结构化字段入库（见下方）。投递/写入失败时，同一条记录会**额外**以 **info** 打到控制台，避免静默丢审计。

默认（控制台）——可省略 **`output`** 或只写 **`provider`**：

```yaml
json_audit:
  enable: true
  output:
    provider: console
```

本地文件：

```yaml
json_audit:
  enable: true
  output:
    provider: file
    file:
      path: /var/log/api-gateway/json-audit.ndjson
```

HTTP 采集：

```yaml
json_audit:
  enable: true
  output:
    provider: http
    http:
      url: https://logs.example.com/ingest/json-audit
      method: POST
      headers:
        Authorization: Bearer your-ingest-token
      timeout_seconds: 8
```

写入数据库（PostgreSQL / MySQL / SQLite）：

```yaml
json_audit:
  enable: true
  output:
    provider: database
    database:
      dsn: postgres://postgres:secret@127.0.0.1:5432/apigw_audit?sslmode=disable
```

```yaml
json_audit:
  enable: true
  output:
    provider: database
    database:
      dsn: mysql://root:secret@127.0.0.1:3306/apigw_audit?charset=utf8mb4&parseTime=True&loc=Local
```

```yaml
json_audit:
  enable: true
  output:
    provider: database
    database:
      dsn: sqlite:///var/lib/api-gateway/json-audit.sqlite
```

推荐使用 URL 风格 DSN（如 `postgres://...`、`mysql://...`、`sqlite:///...`）；URL DSN 下可省略 `engine`。历史 DSN 写法仍保持兼容。

完整数据库字段配置（不使用 `dsn`）：

```yaml
json_audit:
  enable: true
  output:
    provider: database
    database:
      engine: postgres
      host: 127.0.0.1
      port: 5432
      username: postgres
      password: secret
      db: apigw_audit
```

```yaml
json_audit:
  enable: true
  output:
    provider: database
    database:
      engine: mysql
      host: 127.0.0.1
      port: 3306
      username: root
      password: secret
      db: apigw_audit
```

```yaml
json_audit:
  enable: true
  output:
    provider: database
    database:
      engine: sqlite
      db: /var/lib/api-gateway/json-audit.sqlite
```

> 注意：`json_audit.output.provider: database` 必须在 `json_audit.output.database` 内配置数据库连接信息；不支持复用顶层 `database` 作为兜底。

数据库落表（`json_audit_records`）为结构化字段，除基础审计字段外还会提取并保存鉴权相关数据：

- `username`、`password`：来自 `Authorization: Basic ...`（可解码时）。
- `token`：来自 `Authorization: Bearer ...`。
- `authorization`：保存原始 `Authorization` 头值。
- `x_api_key`：来自 `X-API-Key` 头。
- `client_id`、`client_secret`：优先取 `X-Client-ID` / `X-Client-Secret`，仅当对应 header 缺失时回退 query 参数 `client_id` / `client_secret`；不再从 body 提取。

仅某一路由写入单独文件（示例：账单前缀）：

```yaml
routes:
  - path: /billing
    json_audit:
      enable: true
      output:
        provider: file
        file:
          path: /var/log/api-gateway/billing-audit.ndjson
    backend:
      service:
        name: billing
```

#### `max_body_bytes`

限制从**客户端请求体**与**上游响应体**为审计最多读取的字节数；超出部分截断，审计 JSON 里的 **`request_truncated`** / **`response_truncated`** 会标 **`true`**。内存敏感环境可调小；默认 **1 MiB**。

```yaml
json_audit:
  enable: true
  max_body_bytes: 262144   # 256 KiB
```

#### `sample_rate`

在 **include/exclude** 路径规则通过后，对剩余请求按 **`sample_rate`** 做随机抽样，取值 **`(0,1]`**；**`1`** 表示尽量全量；**`≤0`** 在插件内部会按「全量参与抽样」处理（详见下文 [抽样](#抽样)）。**`0.1`** 约等于 **10%** 请求。

```yaml
json_audit:
  enable: true
  sample_rate: 0.1
```

#### `sniff_json`

**`true`**（默认）时，即使 **`Content-Type`** 不像 JSON，只要去掉空白后的 body 满足 **`json.Valid`**，仍可能判为「JSON 类」并写审计。若只希望 **Content-Type 声明为 JSON** 时才审计，设为 **`false`**。

```yaml
json_audit:
  enable: true
  sniff_json: false
```

#### `decompress_gzip`

**`true`**（默认）且响应 **`Content-Encoding`** 含 **gzip** 时，会在 **`max_body_bytes`** 内解压后再做 JSON 判定与 **`response.body`** 记录。若上游从不 gzip JSON，可设为 **`false`** 省 CPU。

```yaml
json_audit:
  enable: true
  decompress_gzip: false
```

#### `include_paths` 与 `exclude_paths`

均按网关 **`ctx.Path`** 做**前缀匹配**。**`include_paths` 非空**时，路径须以其中**至少一条**前缀开头；**`exclude_paths`** 在 include 之后判断，命中则**跳过**审计。常用 **`exclude_paths`** 排除 **`/health`**、监控、静态资源或大体积接口。

```yaml
json_audit:
  enable: true
  include_paths:
    - /api/
  exclude_paths:
    - /health
    - /metrics
    - /static/
```

#### `redact`（嵌套）

**`redact.enable`** 控制是否脱敏；显式配置优先。若省略，默认仅 **`provider=console`** 开启，**`file/http/database`** 默认关闭。设为 **`false`** 时，审计记录中的敏感 **HTTP 头**、**query**、**JSON 字段**均以**明文**写出——仅限受控环境。

**`redact.keys`**：键名**不区分大小写**，JSON 任意层级与 query 参数名命中则值替换为 **`"[REDACTED]"`**。脱敏**开启**且 **`keys`** 为空时，使用内置敏感键列表（含 **`password`**、**`token`** 等）；敏感头（如 **`Authorization`**）仍按内置规则掩码。

```yaml
json_audit:
  enable: true
  redact:
    keys:
      - password
      - national_id
      - bank_account
```

关闭脱敏：

```yaml
json_audit:
  enable: true
  redact:
    enable: false
```

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

当脱敏开启（**`redact.enable=true`**，或未显式配置且 **`provider=console`**）时，若 body 可被解析为 JSON，则对对象键名做匹配（**不区分大小写**，**任意层级**），命中则将值替换为 **`"[REDACTED]"`**。

**`redact.keys` 为空**时使用内置默认值，包含：

`password`、`passwd`、`secret`、`token`、`authorization`、`api_key`、`apikey`、`access_token`、`refresh_token`。

敏感 **HTTP 头**（如 **`Authorization`**、**`Cookie`**）在脱敏开启时由插件内置规则掩码；**query** 参数名命中 **`redact.keys`**（或内置列表）时掩码。若正文无法解析为 JSON，仍写入 **`request.body`** / **`response.body`**，一般为**字符串**片段。脱敏关闭（显式关闭，或未显式配置且 provider 非 console）时，JSON 解析结果与请求头可保持明文。

## 日志字段说明

每条审计日志为 **一行 JSON**。在 **`output.provider: console`**（默认）下通过 **`ctx.Logger.Infof`**（**info**）输出；**`file`** / **`http`** 见上文 **配置说明（`json_audit`）**。

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
| **`headers`** | 请求头，`map[string][]string`；脱敏开启时内置敏感头为 **`["[REDACTED]"]`**；脱敏关闭（显式关闭或 provider 默认关闭）时为原始值。 |
| **`query`** | URL 查询参数，`map[string][]string`；键名命中 **`redact.keys`**（或内置）且脱敏开启时掩码。 |
| **`params`** | 路由参数 **`ctx.Params().ToMap()`**，无则为空对象 **`{}`**。 |
| **`body`** | 请求体：脱敏开启时为键脱敏后的 JSON 或字符串；关闭时为未脱敏解析结果或字符串。 |

**`response` 对象**

| 字段 | 含义 |
| --- | --- |
| **`status`** | 上游 HTTP 状态码。 |
| **`body`** | 响应体：脱敏开启时为键脱敏后的 JSON 或字符串；关闭时为未脱敏解析结果或字符串。 |

使用 **`output.provider: console`** 时可接入现有日志采集；**`output.file.path`** 可落地为 **NDJSON** 再由采集器读取。涉密环境勿依赖默认脱敏作为唯一防护。

## 示例：配置与日志

### 场景说明

客户端 **`POST /api/v1/login`**，请求体为 JSON（含账号口令）；上游返回 **`200`**，响应体为 JSON（含会话 **`token`**）。网关开启 **`json_audit`**，并在 **`provider=console`**（或显式 `redact.enable=true`）下使用默认脱敏键（含 **`password`**、**`token`** 等）。

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
  redact:
    keys:
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

## 限制与注意

- **全量缓冲**：请求与响应会在内存中缓冲至 **`max_body_bytes`**；流式、超大 body 会占用内存或被截断。
- **`Content-Type` 声明为 JSON 但正文非法**：仍可能按「JSON 类」处理，**`request.body`** / **`response.body`** 可能为**字符串**，且无键级脱敏。
- **抽样**为随机过程，短时间窗口内比例会有波动。

## 相关链接

- [插件总览](./)
- [配置说明](/zh/guide/configuration)
