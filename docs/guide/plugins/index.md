# Plugins

API Gateway exposes a dedicated **plugins** area: each built-in capability lives in `github.com/go-zoox/api-gateway/plugin/…` with its own package and lifecycle hooks. This section describes the shared `Plugin` interface, request/response lifecycle, and links to **one page per plugin**.

## Built-in plugins

| Plugin | Package | Enabled when |
| --- | --- | --- |
| [Base URI](./base-uri) | `plugin/baseuri` | YAML `baseuri` is non-empty |
| [IP policy](./ip-policy) | `plugin/ippolicy` | Global `ip_policy.enable` or any route has `ip_policy.enable` |
| [CORS](./cors) | `plugin/cors` | Global `cors.enable` or any route has `cors.enable` |
| [Rate limiting](./rate-limit) | `plugin/ratelimit` | Global `rate_limit.enable` or any route has `rate_limit.enable` |
| [JSON audit](./json-audit) | `plugin/jsonaudit` | Top-level `json_audit.enable` or any route has `json_audit.enable` |

## Plugin interface

Plugins implement:

```go
type Plugin interface {
    Prepare(app *zoox.Application, cfg *config.Config) error
    OnRequest(ctx *zoox.Context, req *http.Request) error
    OnResponse(ctx *zoox.Context, res *http.Response) error
}
```

`Prepare` runs once at startup. `OnRequest` runs before the request is forwarded to a backend; `OnResponse` runs after the upstream response is received.

## Lifecycle

1. **Prepare** — Register middleware, open connections, read config.
2. **OnRequest** — Authentication, rewriting, rate limits, logging.
3. **OnResponse** — Response headers/body tweaks, metrics.

Returning a non-nil error from `OnRequest` or `OnResponse` stops processing for that request (typically returning an HTTP error to the client).

## Custom plugins

1. Embed `plugin.Plugin` (optional) and implement all three methods.
2. Attach the plugin when building the gateway:

```go
import (
    "github.com/go-zoox/api-gateway/config"
    "github.com/go-zoox/api-gateway/core"
    "github.com/go-zoox/api-gateway/plugin"
)

app, err := core.New(version, cfg)
if err != nil {
    return err
}

app.Plugin(&MyPlugin{})
return app.Run()
```

### Example skeleton

```go
type MyPlugin struct {
    plugin.Plugin
}

func (p *MyPlugin) Prepare(app *zoox.Application, cfg *config.Config) error {
    app.Use(func(ctx *zoox.Context) {
        ctx.Next()
    })
    return nil
}

func (p *MyPlugin) OnRequest(ctx *zoox.Context, req *http.Request) error {
    req.Header.Set("X-Example", "1")
    return nil
}

func (p *MyPlugin) OnResponse(ctx *zoox.Context, res *http.Response) error {
    return nil
}
```

## Best practices

1. Keep `OnRequest`/`OnResponse` fast; offload heavy work to workers if needed.
2. Share mutable state only with proper synchronization.
3. Use `ctx.Logger` for consistent logs.
4. Prefer configuration via `cfg` and typed structs under `config` / routes.

## Next steps

- [Base URI](./base-uri) — Strip or require a URL prefix for all routes.
- [IP policy](./ip-policy) — Allow/deny CIDRs; trusted forwarders and `X-Forwarded-For`.
- [CORS](./cors) — Preflight and response headers for browser cross-origin access.
- [Rate limiting](./rate-limit) — Token bucket, leaky bucket, fixed window; memory or Redis.
- [JSON audit](./json-audit) — Log JSON-like responses with paired request bodies for audits.
- [Configuration](/guide/configuration) — YAML structure and globals.
- [API reference](/api/plugin) — `plugin` package overview.
