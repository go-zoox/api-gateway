# CORS plugin

Package: `github.com/go-zoox/api-gateway/plugin/cors`

The CORS plugin is registered when **global** `cors.enable` is true **or** any route sets `cors.enable`. It does two things:

1. **Preflight** (`OPTIONS` with `Origin` and `Access-Control-Request-Method`): the gateway returns **204 No Content** with CORS headers and does not forward the request to a backend.
2. **Normal requests**: the gateway adds CORS headers in `OnResponse` to the upstream response (e.g. `Access-Control-Allow-Origin`).

`allow_origins: ["*"]` is only valid when `allow_credentials` is **false** (the plugin fails at `Prepare` otherwise).

## Configuration

### Global

```yaml
cors:
  enable: true
  allow_origins:
    - "https://app.example.com"
  allow_methods:
    - GET
    - POST
    - PUT
  allow_headers:
    - "*"
  expose_headers:
    - X-Request-Id
  allow_credentials: true
  max_age: 86400
```

### Per route

Enable `cors` on a route. Empty arrays on the route **inherit** from the global block; set fields you want to override.

## Field reference

| Field | Type | Description |
| --- | --- | --- |
| `enable` | bool | Turn CORS on. |
| `allow_origins` | list of strings | `*` (no credentials) or exact origins. |
| `allow_methods` | list | Defaults to common REST methods if omitted. |
| `allow_headers` | list | Default `["*"]` if omitted. |
| `expose_headers` | list | `Access-Control-Expose-Headers`. |
| `allow_credentials` | bool | If true, origins must be explicit (no `*`). |
| `max_age` | int (seconds) | `Access-Control-Max-Age` for preflight. |

## See also

- Full example: `docs/examples/cors.yaml` in this repository.
- [IP policy](./ip-policy) — Run before the browser hits CORS on untrusted networks.
- [Configuration](/guide/configuration) — Route structure.
