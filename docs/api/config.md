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
    Storage   string            `config:"storage,default=memory"`
    KeyType   string            `config:"key_type,default=ip"`
    KeyHeader string            `config:"key_header"`
    Limit     int64             `config:"limit"`
    Window    int64             `config:"window"`
    Burst     int64             `config:"burst"`
    Message   string            `config:"message,default=Too Many Requests"`
    Headers   map[string]string `config:"headers"`
}
```

### Fields

- `Enable` (bool): Enable rate limiting. Default: `false`
- `Algorithm` (string): Rate limiting algorithm. Options: `token-bucket`, `leaky-bucket`, `fixed-window`. Default: `token-bucket`
- `Storage` (string): Storage backend. Options: `memory`, `redis`. Default: `memory`
- `KeyType` (string): Key extraction type. Options: `ip`, `user`, `apikey`, `header`. Default: `ip`
- `KeyHeader` (string): Header name when `KeyType` is `header`
- `Limit` (int64): Maximum number of requests allowed in the time window
- `Window` (int64): Time window in seconds
- `Burst` (int64): Burst capacity (only for token-bucket algorithm)
- `Message` (string): Error message when rate limit is exceeded. Default: `"Too Many Requests"`
- `Headers` (map[string]string): Custom headers to include in rate limit responses

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
