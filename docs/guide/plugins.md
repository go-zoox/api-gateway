# Plugins

API Gateway provides a plugin system for extending functionality.

## Plugin Interface

Plugins implement the `Plugin` interface:

```go
type Plugin interface {
    // Prepare is called during gateway initialization
    Prepare(app *zoox.Application, cfg *config.Config) error
    
    // OnRequest is called before forwarding the request
    OnRequest(ctx *zoox.Context, req *http.Request) error
    
    // OnResponse is called after receiving the response
    OnResponse(ctx *zoox.Context, res *http.Response) error
}
```

## Built-in Plugins

### BaseURI Plugin

The BaseURI plugin filters requests by a base URI prefix. This is automatically enabled when `baseuri` is configured:

```yaml
baseuri: /v1
```

This ensures all requests must start with `/v1`. Requests that don't match return a 404 error.

## Custom Plugins

To create a custom plugin:

1. Implement the `Plugin` interface
2. Register the plugin with the gateway

### Example Plugin

```go
package myplugin

import (
    "net/http"
    "github.com/go-zoox/api-gateway/config"
    "github.com/go-zoox/api-gateway/plugin"
    "github.com/go-zoox/zoox"
)

type MyPlugin struct {
    plugin.Plugin
    // Your plugin configuration
}

func (p *MyPlugin) Prepare(app *zoox.Application, cfg *config.Config) error {
    // Initialize your plugin
    // Add middleware, routes, etc.
    app.Use(func(ctx *zoox.Context) {
        // Your middleware logic
        ctx.Next()
    })
    return nil
}

func (p *MyPlugin) OnRequest(ctx *zoox.Context, req *http.Request) error {
    // Modify request before forwarding
    req.Header.Set("X-Custom-Header", "value")
    return nil
}

func (p *MyPlugin) OnResponse(ctx *zoox.Context, res *http.Response) error {
    // Modify response before returning
    res.Header.Set("X-Response-Header", "value")
    return nil
}
```

### Registering Plugins

Plugins are registered programmatically when creating the gateway:

```go
import (
    "github.com/go-zoox/api-gateway/core"
    "github.com/go-zoox/api-gateway/config"
    "myplugin"
)

cfg := &config.Config{
    // ... configuration
}

app, err := core.New(version, cfg)
if err != nil {
    // handle error
}

app.Plugin(&myplugin.MyPlugin{
    // plugin configuration
})

err = app.Run()
```

## Plugin Lifecycle

1. **Prepare**: Called during gateway initialization. Use this to:
   - Register middleware
   - Set up routes
   - Initialize resources

2. **OnRequest**: Called for each request before forwarding. Use this to:
   - Modify request headers
   - Transform request body
   - Add authentication
   - Log requests

3. **OnResponse**: Called for each response after receiving. Use this to:
   - Modify response headers
   - Transform response body
   - Add caching headers
   - Log responses

## Plugin Examples

### Authentication Plugin

```go
func (p *AuthPlugin) OnRequest(ctx *zoox.Context, req *http.Request) error {
    token := req.Header.Get("Authorization")
    if token == "" {
        return fmt.Errorf("unauthorized")
    }
    
    // Validate token
    if !p.validateToken(token) {
        return fmt.Errorf("invalid token")
    }
    
    return nil
}
```

### Logging Plugin

```go
func (p *LoggingPlugin) OnRequest(ctx *zoox.Context, req *http.Request) error {
    p.logger.Infof("Request: %s %s", req.Method, req.URL.Path)
    return nil
}

func (p *LoggingPlugin) OnResponse(ctx *zoox.Context, res *http.Response) error {
    p.logger.Infof("Response: %d", res.StatusCode)
    return nil
}
```

### Rate Limiting Plugin

The Rate Limiting plugin is automatically enabled when rate limit configuration is present in your gateway configuration. It supports:

- **Multiple key types**: IP address, user ID, API key, or custom header
- **Multiple algorithms**: Token bucket, leaky bucket, or fixed window
- **Storage backends**: In-memory (single instance) or Redis (distributed)
- **Global and route-level configuration**: Configure limits globally or per route

#### Configuration

Rate limiting can be configured globally or per route:

```yaml
# Global rate limit
rate_limit:
  enable: true
  algorithm: token-bucket
  storage: redis
  key_type: ip
  limit: 100
  window: 60
  burst: 20
  message: "Rate limit exceeded"

routes:
  - name: User Service
    path: /v1/user
    rate_limit:
      enable: true
      algorithm: token-bucket
      storage: memory
      key_type: user
      limit: 10
      window: 60
      burst: 5
    backend:
      service:
        # ... backend config
```

#### Configuration Options

- `enable` (bool): Enable rate limiting. Default: `false`
- `algorithm` (string): Rate limiting algorithm. Options: `token-bucket`, `leaky-bucket`, `fixed-window`. Default: `token-bucket`
- `storage` (string): Storage backend. Options: `memory`, `redis`. Default: `memory`
- `key_type` (string): Key extraction type. Options: `ip`, `user`, `apikey`, `header`. Default: `ip`
- `key_header` (string): Header name when `key_type` is `header`
- `limit` (int64): Maximum number of requests allowed in the time window
- `window` (int64): Time window in seconds
- `burst` (int64): Burst capacity (only for token-bucket algorithm)
- `message` (string): Error message when rate limit is exceeded. Default: `"Too Many Requests"`
- `headers` (map[string]string): Custom headers to include in rate limit responses

#### Response Headers

When rate limiting is active, the following headers are included in responses:

- `X-RateLimit-Limit`: The rate limit ceiling for the given request
- `X-RateLimit-Remaining`: The number of requests left for the time window
- `X-RateLimit-Reset`: The time at which the current rate limit window resets (Unix timestamp)
- `Retry-After`: The number of seconds to wait before retrying (only on 429 responses)

## Best Practices

1. **Error Handling**: Return errors from `OnRequest` or `OnResponse` to stop processing
2. **Performance**: Keep plugin logic lightweight to avoid impacting gateway performance
3. **Thread Safety**: Ensure plugins are thread-safe if they maintain state
4. **Configuration**: Use the config parameter to access gateway configuration
5. **Logging**: Use the context logger for consistent logging

## Next Steps

- [Configuration](/guide/configuration) - Learn about gateway configuration
- [Examples](/guide/examples) - See plugin examples in action
- [API Reference](/api/plugin) - Complete plugin API documentation
