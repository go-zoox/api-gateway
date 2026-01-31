# Plugin API

The `Plugin` interface allows extending API Gateway functionality.

## Interface Definition

```go
type Plugin interface {
    Prepare(app *zoox.Application, cfg *config.Config) error
    OnRequest(ctx *zoox.Context, req *http.Request) error
    OnResponse(ctx *zoox.Context, res *http.Response) error
}
```

## Methods

### Prepare

```go
Prepare(app *zoox.Application, cfg *config.Config) error
```

Called during gateway initialization. Use this to:
- Register middleware
- Set up routes
- Initialize resources

**Parameters:**
- `app` (*zoox.Application): The Zoox application instance
- `cfg` (*config.Config): Gateway configuration

**Returns:**
- `error`: Error if initialization fails

### OnRequest

```go
OnRequest(ctx *zoox.Context, req *http.Request) error
```

Called before forwarding the request to the backend. Use this to:
- Modify request headers
- Transform request body
- Add authentication
- Log requests

**Parameters:**
- `ctx` (*zoox.Context): Request context
- `req` (*http.Request): HTTP request (can be modified)

**Returns:**
- `error`: Error to stop processing and return error response

### OnResponse

```go
OnResponse(ctx *zoox.Context, res *http.Response) error
```

Called after receiving the response from the backend. Use this to:
- Modify response headers
- Transform response body
- Add caching headers
- Log responses

**Parameters:**
- `ctx` (*zoox.Context): Request context
- `res` (*http.Response): HTTP response (can be modified)

**Returns:**
- `error`: Error to return error response

## Example Implementation

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
    // Plugin configuration
}

func (p *MyPlugin) Prepare(app *zoox.Application, cfg *config.Config) error {
    // Initialize plugin
    app.Use(func(ctx *zoox.Context) {
        // Middleware logic
        ctx.Next()
    })
    return nil
}

func (p *MyPlugin) OnRequest(ctx *zoox.Context, req *http.Request) error {
    // Modify request
    req.Header.Set("X-Custom-Header", "value")
    return nil
}

func (p *MyPlugin) OnResponse(ctx *zoox.Context, res *http.Response) error {
    // Modify response
    res.Header.Set("X-Response-Header", "value")
    return nil
}
```

## Registering Plugins

Plugins are registered programmatically:

```go
app, err := core.New(version, cfg)
if err != nil {
    // handle error
}

app.Plugin(&myplugin.MyPlugin{
    // configuration
})

err = app.Run()
```

## Built-in Plugins

### BaseURI Plugin

The BaseURI plugin is automatically enabled when `baseuri` is configured:

```go
type BaseURI struct {
    plugin.Plugin
    BaseURI string
}
```

This plugin filters requests by base URI prefix.

## Best Practices

1. **Error Handling**: Return errors to stop processing
2. **Performance**: Keep plugin logic lightweight
3. **Thread Safety**: Ensure plugins are thread-safe
4. **Logging**: Use context logger for consistent logging
5. **Configuration**: Access gateway config through `cfg` parameter

## See Also

- [Plugins Guide](/guide/plugins) - Plugin development guide
- [Config API](/api/config) - Configuration API
