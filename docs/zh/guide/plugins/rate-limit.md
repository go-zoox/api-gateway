# 限流插件

包路径：`github.com/go-zoox/api-gateway/plugin/ratelimit`

当**全局** `rate_limit.enable` 为真，或**任意路由**启用 `rate_limit.enable` 时，会注册限流插件。它在 `OnRequest` 阶段执行；超出配额时可返回 **429 Too Many Requests**。

## 能力概览

- **限流维度**：IP（支持 `X-Forwarded-For`、`X-Real-IP`）、用户（Bearer / `X-User-ID`）、API Key（`X-API-Key`、`Authorization: ApiKey …`、`api_key` 查询参数）、**客户端 ID**（`X-Client-ID` 或查询参数 `client_id`）、自定义请求头。
- **算法**：`token-bucket`、`leaky-bucket`、`fixed-window`。
- **计数器**：仅通过 **`Application.Cache()`** 持久化（顶层配置 `cache` 时一般为 Redis；未配置时为 zoox 默认内存 KV）。
- **作用域**：全局默认策略 + **路由级**覆盖。

## 配置示例

### 全局

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

### 路由级

路由上的 `rate_limit` 对匹配路径覆盖全局策略。

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

### 字段说明（`rate_limit`）

YAML 使用 **snake_case**（如 `key_type`）。「**是否必须**」表示要写出一项有效限流策略时是否必须配置该字段；「**默认值**」表示可省略或未写时的取值。**简要说明**列只放摘要；完整用法见下文 [字段详解](#field-details)。

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
<th align="left">字段</th>
<th align="left">是否必须</th>
<th align="left">默认值</th>
<th align="left">简要说明</th>
</tr>
</thead>
<tbody>
<tr>
<td valign="top"><code>limit</code></td>
<td valign="top">是</td>
<td valign="top">—</td>
<td valign="top">单个限流键在一个活跃窗口内允许的最大请求次数。<a href="#field-limit">详细说明</a></td>
</tr>
<tr>
<td valign="top"><code>window</code></td>
<td valign="top">是</td>
<td valign="top">—</td>
<td valign="top">窗口长度（秒），与 <code>limit</code> 共同决定 refill/泄漏/计数节奏。<a href="#field-window">详细说明</a></td>
</tr>
<tr>
<td valign="top"><code>enable</code></td>
<td valign="top">否</td>
<td valign="top"><code>false</code></td>
<td valign="top">在全局或路由范围打开限流插件。<a href="#field-enable">详细说明</a></td>
</tr>
<tr>
<td valign="top"><code>algorithm</code></td>
<td valign="top">否</td>
<td valign="top"><code>token-bucket</code></td>
<td valign="top">使用的限流算法（令牌桶 / 漏桶 / 固定窗口）。<a href="#field-algorithm">详细说明</a></td>
</tr>
<tr>
<td valign="top"><code>key_type</code></td>
<td valign="top">否</td>
<td valign="top"><code>ip</code></td>
<td valign="top">如何从请求中构造限流键（客户端维度）。<a href="#field-key-type">详细说明</a></td>
</tr>
<tr>
<td valign="top"><code>key_header</code></td>
<td valign="top">否</td>
<td valign="top"><em>空</em></td>
<td valign="top"><code>key_type: header</code> 时指定参与键值的请求头名称。<a href="#field-key-header">详细说明</a></td>
</tr>
<tr>
<td valign="top"><code>burst</code></td>
<td valign="top">否</td>
<td valign="top"><code>0</code></td>
<td valign="top">主要针对令牌桶的桶容量；其它算法常忽略。<a href="#field-burst">详细说明</a></td>
</tr>
<tr>
<td valign="top"><code>message</code></td>
<td valign="top">否</td>
<td valign="top"><code>Too Many Requests</code></td>
<td valign="top">返回 429 时的响应正文。<a href="#field-message">详细说明</a></td>
</tr>
<tr>
<td valign="top"><code>headers</code></td>
<td valign="top">否</td>
<td valign="top"><em>空映射</em></td>
<td valign="top">仅在 429 响应上附加的额外 HTTP 头。<a href="#field-headers">详细说明</a></td>
</tr>
</tbody>
</table>
</div>

**书写位置：** 顶层 `rate_limit` 提供全局默认；路由下的 `rate_limit` 覆盖匹配该路由的请求（参见「路由匹配优先级」）。

## 字段详解

<a id="field-details"></a>

以下按字段说明：**含义**、**默认值**、**用法**，并附 **YAML 示例**。

<a id="field-limit"></a>
### `limit`

- **含义：** 在一个策略时间窗口内，对**同一个限流键**（见 `key_type`）允许通过的最大请求次数；计数细节由 `algorithm` 决定。
- **默认值：** 无单独默认值——该字段**必填**。若为 0 或负数，该策略会被跳过（fail-open）。
- **用法：** 可按路由收紧昂贵接口；与 `window` 搭配表示「每 `window` 秒最多 `limit` 次」。
- **示例：** 每个窗口最多 **100** 次（窗口长度由 `window` 给出）：

```yaml
rate_limit:
  enable: true
  limit: 100
  window: 60
```

<a id="field-window"></a>
### `window`

- **含义：** 策略时间窗口长度，单位为**秒**（整数）。令牌桶 refill 约为 `limit/window`；固定窗口按存储滚动；漏桶泄漏率比例与此一致。
- **默认值：** 无——**必填**。若为 0 或负数，该策略被跳过。
- **用法：** 窗口短则对突发反应快；长则更平滑；需与客户端重试、`Retry-After` 预期一致。
- **示例：** `limit: 100`、`window: 60` 表示在 60 秒窗口内计数（具体边界随算法略有差异）：

```yaml
rate_limit:
  enable: true
  limit: 100
  window: 60
```

<a id="field-enable"></a>
### `enable`

- **含义：** 是否在**当前作用域**启用限流：根配置的 `rate_limit`（全局默认）或某条路由下的 `rate_limit`（仅匹配路径）。
- **默认值：** `false`。仅当**全局或至少一条路由**存在 `enable: true` 时才会注册插件。
- **用法：** 可只在敏感路由开启；或全局开启再在个别路由覆盖其它字段。
- **示例：** 仅对某一路由开启（全局可为 `false` 或不写）：

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

- **含义：** 使用的限流算法：`token-bucket`（默认）、`leaky-bucket`、`fixed-window`；概述见 [算法简述](#算法简述)。未知字符串会按 **令牌桶** 处理。
- **默认值：** `token-bucket`（省略时）。
- **用法：** 简单计数用固定窗口；需要 refill + 突发用令牌桶；希望输出更平滑用漏桶。
- **示例：** 按 IP 使用固定窗口计数：

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

- **含义：** 如何从请求构造限流键。取值：`ip`、`user`、`apikey`、`clientid`、`header`；其它字符串按 **`ip`**。
- **默认值：** `ip`。
- **补充：** `ip` 依次用 `X-Forwarded-For`、`X-Real-IP`、`RemoteAddr`；`user` 用 Bearer 或 `X-User-ID`，否则退回 IP；`apikey` 用 `X-API-Key` / `Authorization: ApiKey …` / 查询参数 `api_key`，否则退回 IP；`clientid` 优先 `X-Client-ID`，否则查询参数 `client_id`，否则退回 IP；`header` 配合 `key_header`，头名为空则退回 IP。
- **用法：** 匿名用 `ip`；鉴权后用 `user`/`apikey`；稳定客户端标识用 `clientid`；多租户等用 `header`。
- **示例（API Key）：**

```yaml
rate_limit:
  enable: true
  key_type: apikey
  limit: 1000
  window: 3600
```

- **示例（`clientid`）：**

```yaml
rate_limit:
  enable: true
  key_type: clientid
  limit: 200
  window: 60
```

<a id="field-key-header"></a>
### `key_header`

- **含义：** 当 `key_type: header` 时，填写**请求头名称**（不是取值），限流键会包含该头的值。
- **默认值：** 空。`key_type` 为 `header` 且为空时，**退回按 IP** 提取。
- **用法：** 填写稳定维度，如租户、环境标识。
- **示例：** 每个 **租户 ID**（`X-Tenant-ID`）单独配额：

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

- **含义：** 对 **token-bucket** 表示桶容量（突发上限）；refill 仍为 `limit/window`。≤0 或未写时，容量等同 **`limit`**。漏桶/固定窗口可能忽略。
- **默认值：** `0`（对令牌桶表示「容量用 `limit`」）。
- **用法：** 仅当短时突发需**高于** sustained `limit`/`window` 对应能力时，将 `burst` 设为大于 `limit`。
- **示例：** `limit: 10`、`window: 1` 近似 10 req/s，但桶内最多积攒 **50** 个令牌用于突发：

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

- **含义：** 因限流返回 **HTTP 429** 时的**响应正文**。
- **默认值：** `Too Many Requests`（空串时网关可能仍回退到默认文案，视实现而定）。
- **用法：** 可改为 JSON 或与现有错误体风格统一的字符串。
- **示例：** 返回 JSON 风格错误（整段用 YAML 引号包住）：

```yaml
rate_limit:
  enable: true
  limit: 60
  window: 60
  message: '{"error":"too_many_requests","retry":true}'
```

<a id="field-headers"></a>
### `headers`

- **含义：** 仅在 **429** 时附加的额外**响应头**；与网关自动下发的 `X-RateLimit-*`、`Retry-After`（可写时）并存。
- **默认值：** 空映射，不加额外头。
- **用法：** 适合策略名、文档链接、追踪 id；勿写密钥。
- **示例：** 打上策略标签便于监控区分：

```yaml
rate_limit:
  enable: true
  limit: 100
  window: 60
  headers:
    X-Rate-Policy: standard-tier
```


## 路由匹配优先级

多条路由都配置了限流时：

1. 按路由路径**长度从长到短**排序。
2. 请求路径需与某条路由**完全相同**，或作为**前缀**匹配且前缀后的下一个字符为 `/`（例如 `/users` 匹配 `/users/123`，但不匹配 `/users-extra`）。

更长路径优先于较短前缀（例如 `/api/v1/foo` 优先匹配 `/api/v1` 而非 `/api`）。

## 算法简述

| 算法 | 行为 |
| --- | --- |
| `token-bucket` | 允许最大 `burst` 的突发；按 `limit`/`window`  refill。 |
| `leaky-bucket` | 平滑输出；突发语义与令牌桶不同。 |
| `fixed-window` | 固定窗口计数，具体边界由存储实现决定。 |

## 与 Application.Cache 的关系

限流计数器只通过 **`app.Cache()`**（内部即 `cache.New(&app.Config.Cache)`）读写。未配置顶层 `cache` 时，zoox 默认仍是内存 KV，行为与框架一致。

## 响应头

正常情况下（未触发 429）常见：

- `X-RateLimit-Limit`
- `X-RateLimit-Remaining`
- `X-RateLimit-Reset`（Unix 时间戳）

返回 **429** 时还会设置 `Retry-After`（秒），以及 `headers` 中配置的自定义头。

若算法或存储返回错误，插件会**放行请求**（fail-open）并记录日志。

## 相关链接

- [插件总览](./)
- [配置](/zh/guide/configuration)
- [路由](/zh/guide/routing)
