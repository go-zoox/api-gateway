# Base URI plugin

Package: `github.com/go-zoox/api-gateway/plugin/baseuri`

When the gateway is configured with a global base URI, the Base URI plugin installs middleware that:

1. **Rejects** requests whose path does not begin with the configured prefix (responds with **404 Not Found**).
2. **Strips** that prefix from `ctx.Request.URL.Path` and updates `ctx.Path` so downstream routing matches routes without repeating the prefix.

This matches the usual pattern of exposing the API under `/api`, `/v1`, etc., while keeping route definitions rooted at `/`.

## Configuration

Set the gateway-level field `baseuri` in YAML:

```yaml
baseuri: /v1
```

With `baseuri: /v1`, an incoming request to `GET /v1/users` is normalized to path `/users` before route matching.

The plugin is attached automatically when `baseuri` is non-empty (`github.com/go-zoox/api-gateway/core` prepares it during startup).

## Behaviour details

| Incoming path | Result |
| --- | --- |
| Starts with configured base URI | Prefix removed; request continues |
| Does not start with prefix | **404**, request does not proceed |

Trailing characters after the prefix must align with normal path semantics: only the configured prefix is removed in one chunk.

## Interaction with other plugins

Routing and plugins see the **stripped** path after Base URI runs. Configure route `path` values relative to that stripped path when using a base URI.

## Related

- [Plugins overview](./)
- [Routing](/guide/routing)
- [Configuration](/guide/configuration)
