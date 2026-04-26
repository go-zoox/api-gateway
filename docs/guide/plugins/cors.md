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

**Required?** is whether the field must be set for a *meaningful* CORS block once `cors` is in use. **Default** is the value when the field is omitted (struct zero / normalizer defaults). To actually enable the plugin, set **`enable: true`** on the **global** `cors` block and/or a **route** (see the table for `enable`).

| Field | Type | Required? | Default | Description |
| --- | --- | --- | --- | --- |
| `enable` | bool | No* | `false` | *Must be `true` (globally and/or on a route) for the CORS plugin to register and run. |
| `allow_origins` | list of strings | No | `["*"]` | `*` (no credentials) or exact `Origin` values. Omitted is treated as allow-all (same as `*`) before validation; still incompatible with `allow_credentials: true`. |
| `allow_methods` | list | No | `GET`, `POST`, `PUT`, `PATCH`, `DELETE`, `HEAD`, `OPTIONS` | Used for preflight and `Access-Control-Allow-Methods`. |
| `allow_headers` | list | No | `["*"]` | `Access-Control-Allow-Headers` on preflight. |
| `expose_headers` | list | No | _empty_ | `Access-Control-Expose-Headers` on real responses. |
| `allow_credentials` | bool | No | `false` | If `true`, `allow_origins` must not be `*`. **Prepare** fails on invalid pairings. |
| `max_age` | int (seconds) | No | `0` | `Access-Control-Max-Age` for preflight; `0` means the header is omitted. |

## See also

- Full example: `docs/examples/cors.yaml` in this repository.
- [IP policy](./ip-policy) — Run before the browser hits CORS on untrusted networks.
- [Configuration](/guide/configuration) — Route structure.
