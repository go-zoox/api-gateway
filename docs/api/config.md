# Config API

The `Config` struct defines the main configuration for API Gateway.

## Type Definition

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

## Fields

### Port

```go
Port int64
```

The port the gateway listens on. Default: `8080`

### BaseURI

```go
BaseURI string
```

Base URI prefix for all routes. When set, only requests starting with this prefix are accepted.

### Backend

```go
Backend route.Backend
```

Default backend service. Used when no route matches.

### Routes

```go
Routes []route.Route
```

Array of route definitions. See [Route API](/api/route) for details.

### Cache

```go
Cache Cache
```

Cache configuration. See [Cache Configuration](#cache-configuration).

### HealthCheck

```go
HealthCheck HealthCheck
```

Health check configuration. See [Health Check Configuration](#health-check-configuration).

## Cache Configuration

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

### Fields

- `Host` (string): Redis host. Default: `127.0.0.1`
- `Port` (int64): Redis port. Default: `6379`
- `Username` (string): Redis username (optional)
- `Password` (string): Redis password (optional)
- `DB` (int64): Redis database number. Default: `0`
- `Prefix` (string): Key prefix. Default: `gozoox-api-gateway:`

## Health Check Configuration

```go
type HealthCheck struct {
    Outer HealthCheckOuter `config:"outer"`
    Inner HealthCheckInner `config:"inner"`
}
```

### Outer Health Check

```go
type HealthCheckOuter struct {
    Enable bool   `config:"enable"`
    Path   string `config:"path"`
    Ok     bool   `config:"ok"`
}
```

- `Enable` (bool): Enable external health check endpoint
- `Path` (string): Health check endpoint path. Default: `/healthz`
- `Ok` (bool): Always return OK. Default: `true`

### Inner Health Check

```go
type HealthCheckInner struct {
    Enable   bool `config:"enable"`
    Interval int64 `config:"interval"`
    Timeout  int64 `config:"timeout"`
}
```

- `Enable` (bool): Enable internal service health checks
- `Interval` (int64): Check interval in seconds. Default: `30`
- `Timeout` (int64): Request timeout in seconds. Default: `5`

## Rate Limit Configuration

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

Counters are stored **only** through `zoox.Application.Cache()` (configure top-level `cache` for Redis; otherwise the framework default applies). Same struct is used on the root config and on each route’s `rate_limit` override.

### Fields

YAML keys are **snake_case** (`key_type`, …); the table uses Go struct field names. **Required?** means the field must be set for an effective policy; **Default** applies when omitted. **Summary** is short; meaning, defaults, usage, and **YAML examples** for each key are in the [Rate limiting plugin — Field details](/guide/plugins/rate-limit#field-details) guide.

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
<th align="left">Field</th>
<th align="left">Go type</th>
<th align="left">Required?</th>
<th align="left">Default</th>
<th align="left">Summary</th>
</tr>
</thead>
<tbody>
<tr>
<td valign="top"><code>Limit</code></td>
<td valign="top"><code>int64</code></td>
<td valign="top">Yes</td>
<td valign="top">—</td>
<td valign="top">YAML <code>limit</code>: quota per window. <a href="/guide/plugins/rate-limit#field-limit">Details</a></td>
</tr>
<tr>
<td valign="top"><code>Window</code></td>
<td valign="top"><code>int64</code></td>
<td valign="top">Yes</td>
<td valign="top">—</td>
<td valign="top">YAML <code>window</code>: window length in seconds. <a href="/guide/plugins/rate-limit#field-window">Details</a></td>
</tr>
<tr>
<td valign="top"><code>Enable</code></td>
<td valign="top"><code>bool</code></td>
<td valign="top">No</td>
<td valign="top"><code>false</code></td>
<td valign="top">YAML <code>enable</code>: scope flag. <a href="/guide/plugins/rate-limit#field-enable">Details</a></td>
</tr>
<tr>
<td valign="top"><code>Algorithm</code></td>
<td valign="top"><code>string</code></td>
<td valign="top">No</td>
<td valign="top"><code>token-bucket</code></td>
<td valign="top">YAML <code>algorithm</code>. <a href="/guide/plugins/rate-limit#field-algorithm">Details</a></td>
</tr>
<tr>
<td valign="top"><code>KeyType</code></td>
<td valign="top"><code>string</code></td>
<td valign="top">No</td>
<td valign="top"><code>ip</code></td>
<td valign="top">YAML <code>key_type</code>. <a href="/guide/plugins/rate-limit#field-key-type">Details</a></td>
</tr>
<tr>
<td valign="top"><code>KeyHeader</code></td>
<td valign="top"><code>string</code></td>
<td valign="top">No</td>
<td valign="top"><em>(empty)</em></td>
<td valign="top">YAML <code>key_header</code>. <a href="/guide/plugins/rate-limit#field-key-header">Details</a></td>
</tr>
<tr>
<td valign="top"><code>Burst</code></td>
<td valign="top"><code>int64</code></td>
<td valign="top">No</td>
<td valign="top"><code>0</code></td>
<td valign="top">YAML <code>burst</code>. <a href="/guide/plugins/rate-limit#field-burst">Details</a></td>
</tr>
<tr>
<td valign="top"><code>Message</code></td>
<td valign="top"><code>string</code></td>
<td valign="top">No</td>
<td valign="top"><code>Too Many Requests</code></td>
<td valign="top">YAML <code>message</code>. <a href="/guide/plugins/rate-limit#field-message">Details</a></td>
</tr>
<tr>
<td valign="top"><code>Headers</code></td>
<td valign="top"><code>map[string]string</code></td>
<td valign="top">No</td>
<td valign="top"><em>(empty)</em></td>
<td valign="top">YAML <code>headers</code>. <a href="/guide/plugins/rate-limit#field-headers">Details</a></td>
</tr>
</tbody>
</table>
</div>

## Example

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
        // ... routes
    },
}
```

## See Also

- [Route API](/api/route) - Route configuration
- [Configuration Guide](/guide/configuration) - Configuration guide
