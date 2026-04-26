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

| Field | Type | Description |
| --- | --- | --- |
| `enable` | bool | Turn the policy on. |
| `allow` | list of CIDRs | If non-empty, client must match one. |
| `deny` | list of CIDRs | Blocked clients get 403. |
| `trusted_proxies` | list of CIDRs | When the direct peer is in this set, trust `X-Forwarded-For` (first) / `X-Real-IP`. |
| `message` | string | Response body for 403. |

CIDRs may be written as `192.0.2.1/32` or a single address without a mask (`192.0.2.1` means `/32` or `/128`).

## See also

- Full example: `docs/examples/ip-policy.yaml` in this repository.
- [CORS](./cors) — Often used after IP allowlists in the chain.
- [Configuration](/guide/configuration) — Top-level and route structure.
