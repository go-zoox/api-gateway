# 配置 API

`Config` 结构体定义了 API Gateway 的主配置。

## 类型定义

```go
type Config struct {
    Port     int64          `config:"port"`
    BaseURI  string         `config:"baseuri"`
    Backend  route.Backend  `config:"backend"`
    Routes   []route.Route  `config:"routes"`
    Cache    Cache          `config:"cache"`
    HealthCheck HealthCheck `config:"healthcheck"`
    RateLimit RateLimit     `config:"rate_limit"`
}
```

## 字段

### Port

```go
Port int64
```

网关监听的端口。默认值：`8080`

### BaseURI

```go
BaseURI string
```

所有路由的基础 URI 前缀。设置后，只接受以此前缀开头的请求。

### Backend

```go
Backend route.Backend
```

默认后端服务。当没有路由匹配时使用。

### Routes

```go
Routes []route.Route
```

路由定义数组。有关详细信息，请参阅[路由 API](/zh/api/route)。

### Cache

```go
Cache Cache
```

缓存配置。请参阅[缓存配置](#缓存配置)。

### HealthCheck

```go
HealthCheck HealthCheck
```

健康检查配置。请参阅[健康检查配置](#健康检查配置)。

### RateLimit

```go
RateLimit RateLimit
```

限流策略（根配置与各路由 `rate_limit` 共用同一结构）。计数器仅通过 `zoox.Application.Cache()` 持久化；字段语义、默认值、用法与 **YAML 示例**见[限流插件 — 字段详解](/zh/guide/plugins/rate-limit#field-details)。亦可直接查看[限流配置](#限流配置)下的字段表。

## 缓存配置

```go
type Cache struct {
    Host     string `config:"host"`
    Port     int64  `config:"port"`
    Username string `config:"username"`
    Password string `config:"password"`
    DB       int64  `config:"db"`
    Prefix   string `config:"prefix"`
}
```

### 字段

- `Host` (string): Redis 主机。默认值：`127.0.0.1`
- `Port` (int64): Redis 端口。默认值：`6379`
- `Username` (string): Redis 用户名（可选）
- `Password` (string): Redis 密码（可选）
- `DB` (int64): Redis 数据库编号。默认值：`0`
- `Prefix` (string): 键前缀。默认值：`gozoox-api-gateway:`

## 健康检查配置

```go
type HealthCheck struct {
    Outer HealthCheckOuter `config:"outer"`
    Inner HealthCheckInner `config:"inner"`
}
```

### 外部健康检查

```go
type HealthCheckOuter struct {
    Enable bool   `config:"enable"`
    Path   string `config:"path"`
    Ok     bool   `config:"ok"`
}
```

- `Enable` (bool): 启用外部健康检查端点
- `Path` (string): 健康检查端点路径。默认值：`/healthz`
- `Ok` (bool): 始终返回 OK。默认值：`true`

### 内部健康检查

```go
type HealthCheckInner struct {
    Enable   bool `config:"enable"`
    Interval int64 `config:"interval"`
    Timeout  int64 `config:"timeout"`
}
```

- `Enable` (bool): 启用内部服务健康检查
- `Interval` (int64): 检查间隔（秒）。默认值：`30`
- `Timeout` (int64): 请求超时（秒）。默认值：`5`

## 限流配置

```go
type RateLimit struct {
    Enable    bool              `config:"enable"`
    Algorithm string            `config:"algorithm,default=token-bucket"`
    KeyType   string            `config:"key_type,default=ip"`
    KeyHeader string            `config:"key_header"`
    Limit     int64             `config:"limit"`
    Window    int64             `config:"window"`
    Burst     int64             `config:"burst"`
    Message   string            `config:"message,default=Too Many Requests"`
    Headers   map[string]string `config:"headers"`
}
```

计数器**仅**通过 `zoox.Application.Cache()` 存储（顶层配置 `cache` 指向 Redis 等；未配置时使用框架默认）。根配置的 `rate_limit` 与各路由上的 `rate_limit` 覆盖项均使用上述结构体。

### 字段

YAML 键为 **snake_case**（如 `key_type`）；下表列为 Go 字段名。**是否必须**表示一项有效策略是否必填该字段；**默认值**表示可省略时的取值。**简要说明**仅作索引；各键的**含义、默认值、用法与 YAML 示例**见 [限流插件 — 字段详解](/zh/guide/plugins/rate-limit#field-details)。

<div style="overflow-x:auto">
<table style="table-layout:fixed;width:100%;max-width:56rem;border-collapse:collapse">
<colgroup>
<col style="width:6.5rem" />
<col style="width:9rem" />
<col style="width:5rem" />
<col style="width:6rem" />
<col style="width:20rem" />
</colgroup>
<thead>
<tr>
<th align="left">字段</th>
<th align="left">Go 类型</th>
<th align="left">是否必须</th>
<th align="left">默认值</th>
<th align="left">简要说明</th>
</tr>
</thead>
<tbody>
<tr>
<td valign="top"><code>Limit</code></td>
<td valign="top"><code>int64</code></td>
<td valign="top">是</td>
<td valign="top">—</td>
<td valign="top">YAML <code>limit</code>，每窗口配额。<a href="/zh/guide/plugins/rate-limit#field-limit">详细说明</a></td>
</tr>
<tr>
<td valign="top"><code>Window</code></td>
<td valign="top"><code>int64</code></td>
<td valign="top">是</td>
<td valign="top">—</td>
<td valign="top">YAML <code>window</code>，窗口长度（秒）。<a href="/zh/guide/plugins/rate-limit#field-window">详细说明</a></td>
</tr>
<tr>
<td valign="top"><code>Enable</code></td>
<td valign="top"><code>bool</code></td>
<td valign="top">否</td>
<td valign="top"><code>false</code></td>
<td valign="top">YAML <code>enable</code>，作用域开关。<a href="/zh/guide/plugins/rate-limit#field-enable">详细说明</a></td>
</tr>
<tr>
<td valign="top"><code>Algorithm</code></td>
<td valign="top"><code>string</code></td>
<td valign="top">否</td>
<td valign="top"><code>token-bucket</code></td>
<td valign="top">YAML <code>algorithm</code>。<a href="/zh/guide/plugins/rate-limit#field-algorithm">详细说明</a></td>
</tr>
<tr>
<td valign="top"><code>KeyType</code></td>
<td valign="top"><code>string</code></td>
<td valign="top">否</td>
<td valign="top"><code>ip</code></td>
<td valign="top">YAML <code>key_type</code>。<a href="/zh/guide/plugins/rate-limit#field-key-type">详细说明</a></td>
</tr>
<tr>
<td valign="top"><code>KeyHeader</code></td>
<td valign="top"><code>string</code></td>
<td valign="top">否</td>
<td valign="top"><em>（空）</em></td>
<td valign="top">YAML <code>key_header</code>。<a href="/zh/guide/plugins/rate-limit#field-key-header">详细说明</a></td>
</tr>
<tr>
<td valign="top"><code>Burst</code></td>
<td valign="top"><code>int64</code></td>
<td valign="top">否</td>
<td valign="top"><code>0</code></td>
<td valign="top">YAML <code>burst</code>。<a href="/zh/guide/plugins/rate-limit#field-burst">详细说明</a></td>
</tr>
<tr>
<td valign="top"><code>Message</code></td>
<td valign="top"><code>string</code></td>
<td valign="top">否</td>
<td valign="top"><code>Too Many Requests</code></td>
<td valign="top">YAML <code>message</code>。<a href="/zh/guide/plugins/rate-limit#field-message">详细说明</a></td>
</tr>
<tr>
<td valign="top"><code>Headers</code></td>
<td valign="top"><code>map[string]string</code></td>
<td valign="top">否</td>
<td valign="top"><em>（空）</em></td>
<td valign="top">YAML <code>headers</code>。<a href="/zh/guide/plugins/rate-limit#field-headers">详细说明</a></td>
</tr>
</tbody>
</table>
</div>

## 示例

```go
cfg := &config.Config{
    Port: 8080,
    BaseURI: "/v1",
    Cache: config.Cache{
        Host: "127.0.0.1",
        Port: 6379,
        DB: 0,
    },
    HealthCheck: config.HealthCheck{
        Outer: config.HealthCheckOuter{
            Enable: true,
            Path: "/healthz",
        },
        Inner: config.HealthCheckInner{
            Enable: true,
            Interval: 30,
            Timeout: 5,
        },
    },
    Routes: []route.Route{
        // ... 路由
    },
}
```

## 另请参阅

- [路由 API](/zh/api/route) - 路由配置
- [配置指南](/zh/guide/configuration) - 配置指南
- [限流插件](/zh/guide/plugins/rate-limit) - 限流字段详解与 YAML 示例
