# IP policy plugin

Package: `github.com/go-zoox/api-gateway/plugin/ippolicy`

The IP policy plugin is registered when **global** `ip_policy.enable` is true **or** any route sets `ip_policy.enable`. It runs in **HTTP middleware** before the reverse proxy, so it applies to all methods (including `OPTIONS` for CORS preflight).

## Behaviour

- **Deny** CIDRs are checked first. A matching client IP receives **403 Forbidden** with the configured `message` body.
- If **allow** is non-empty, the client IP must also match at least one allow CIDR. If allow is empty, only deny is enforced (permissive default).
- **Trusted proxies**: if empty, the gateway **only** uses the direct TCP peer address (`RemoteAddr`) and **ignores** `X-Forwarded-For` / `X-Real-IP`. If non-empty, when the direct peer is inside one of these CIDRs, the **first** hop in `X-Forwarded-For` is treated as the client (same idea as common load balancer deployments).

## Configuration

### Global

```yaml
ip_policy:
  enable: true
  allow:
    - 10.0.0.0/8
    - 2001:db8::/32
  deny:
    - 198.51.100.0/24
  trusted_proxies:
    - 10.0.0.0/8
  message: "Forbidden"
```

### Per route

Set `ip_policy.enable` on a route. Route **allow**, **deny**, **trusted_proxies**, and **message** are merged with the global block: empty route lists inherit from global; non-empty **deny** on a route is **appended** to global deny. **Allow** on a route replaces the global allow list for that route when the route’s allow is non-empty.

## Field reference

**Required?** is whether the field must be set for a *meaningful* IP policy. **Default** is the value when the field is omitted. To actually enable the plugin, set **`enable: true`** on the **global** `ip_policy` block and/or a **route** (see the table for `enable`).

| Field | Type | Required? | Default | Description |
| --- | --- | --- | --- | --- |
| `enable` | bool | No* | `false` | *Must be `true` (globally and/or on a route) for the IP policy plugin to register and run. |
| `allow` | list of CIDRs | No | _empty_ | If non-empty, client must match at least one; if empty, no allowlist is applied (only `deny` matters). |
| `deny` | list of CIDRs | No | _empty_ | If non-empty, matching clients receive 403. |
| `trusted_proxies` | list of CIDRs | No | _empty_ | If empty, only `RemoteAddr` is used; if non-empty, XFF is trusted when the direct peer is in this set. |
| `message` | string | No | `Forbidden` | Response body for HTTP 403 when blocked. |

CIDRs may be written as `192.0.2.1/32` or a single address without a mask (`192.0.2.1` means `/32` or `/128`).

## See also

- Full example: `docs/examples/ip-policy.yaml` in this repository.
- [CORS](./cors) — Often used after IP allowlists in the chain.
- [Configuration](/guide/configuration) — Top-level and route structure.
