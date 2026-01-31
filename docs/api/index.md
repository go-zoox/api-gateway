# API Reference

Complete API reference for API Gateway configuration and programmatic usage.

## Configuration API

- [Config](/api/config) - Main configuration structure
- [Route](/api/route) - Route configuration
- [Plugin](/api/plugin) - Plugin interface

## Programmatic Usage

API Gateway can be used programmatically in Go applications:

```go
import (
    "github.com/go-zoox/api-gateway/core"
    "github.com/go-zoox/api-gateway/config"
)

cfg := &config.Config{
    Port: 8080,
    Routes: []route.Route{
        {
            Name: "api",
            Path: "/api",
            Backend: route.Backend{
                Service: service.Service{
                    Name: "api.example.com",
                    Port: 443,
                    Protocol: "https",
                },
            },
        },
    },
}

app, err := core.New("1.4.5", cfg)
if err != nil {
    // handle error
}

err = app.Run()
```

## Type Definitions

All configuration types are defined in the following packages:

- `github.com/go-zoox/api-gateway/config` - Main configuration
- `github.com/go-zoox/api-gateway/core/route` - Route definitions
- `github.com/go-zoox/api-gateway/core/service` - Service definitions
- `github.com/go-zoox/api-gateway/plugin` - Plugin interface

## See Also

- [Configuration Guide](/guide/configuration) - Configuration guide
- [Examples](/guide/examples) - Configuration examples
