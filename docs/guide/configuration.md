# Configuration

API Gateway uses YAML configuration files. This document describes all available configuration options.

## Configuration File Structure

```yaml
version: v1              # Configuration version
port: 8080               # Gateway listening port
baseuri: /v1             # Base URI prefix (optional)

cache:                   # Cache configuration (optional)
  engine: redis
  host: 127.0.0.1
  port: 6379
  password: ""
  db: 0
  prefix: "gozoox-api-gateway:"

healthcheck:             # Health check configuration
  outer:                 # External health check
    enable: true
    path: /healthz
    ok: true
  inner:                 # Internal service health check
    enable: true
    interval: 30         # Check interval in seconds
    timeout: 5           # Timeout in seconds

backend:                 # Default backend (optional)
  service:
    protocol: https
    name: example.com
    port: 443

routes:                  # Route definitions
  - name: route-name
    path: /api
    path_type: prefix    # prefix or regex
    backend:
      service:
        protocol: https
        name: backend.example.com
        port: 443
        request:
          path:
            disable_prefix_rewrite: false
            rewrites:
              - "^/api/(.*):/$1"
          headers:
            X-Custom-Header: value
          query:
            key: value
        response:
          headers:
            X-Response-Header: value
        auth:
          type: bearer
          token: your-token
```

## Configuration Fields

### Top Level

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `version` | string | Yes | - | Configuration version (currently `v1`) |
| `port` | int | No | 8080 | Port the gateway listens on |
| `baseuri` | string | No | - | Base URI prefix for all routes |
| `cache` | object | No | - | Cache configuration |
| `healthcheck` | object | No | - | Health check configuration |
| `backend` | object | No | - | Default backend service |
| `routes` | array | No | [] | Route definitions |

### Cache Configuration

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `engine` | string | No | redis | Cache engine (currently only `redis`) |
| `host` | string | No | 127.0.0.1 | Redis host |
| `port` | int | No | 6379 | Redis port |
| `password` | string | No | - | Redis password |
| `db` | int | No | 0 | Redis database number |
| `prefix` | string | No | gozoox-api-gateway: | Key prefix |

### Health Check Configuration

#### Outer Health Check

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `enable` | bool | No | false | Enable external health check endpoint |
| `path` | string | No | /healthz | Health check endpoint path |
| `ok` | bool | No | true | Always return OK (skip actual checks) |

#### Inner Health Check

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `enable` | bool | No | false | Enable internal service health checks |
| `interval` | int | No | 30 | Check interval in seconds |
| `timeout` | int | No | 5 | Request timeout in seconds |

### Route Configuration

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `name` | string | Yes | - | Route name (for logging) |
| `path` | string | Yes | - | Path pattern to match |
| `path_type` | string | No | prefix | Match type: `prefix` or `regex` |
| `backend` | object | Yes | - | Backend service configuration |

### Service Configuration

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `protocol` | string | No | http | Protocol: `http` or `https` |
| `name` | string | Yes | - | Service hostname or IP |
| `port` | int | No | 80 | Service port |
| `request` | object | No | - | Request transformation |
| `response` | object | No | - | Response transformation |
| `auth` | object | No | - | Authentication configuration |
| `health_check` | object | No | - | Service-specific health check |

### Request Configuration

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `path` | object | No | - | Path rewriting configuration |
| `headers` | map | No | - | Additional request headers |
| `query` | map | No | - | Additional query parameters |

#### Path Rewriting

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `disable_prefix_rewrite` | bool | No | false | Disable automatic prefix removal |
| `rewrites` | array | No | [] | Path rewrite rules (format: `pattern:replacement`) |

### Response Configuration

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `headers` | map | No | - | Additional response headers |

### Authentication Configuration

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `type` | string | No | - | Auth type: `basic`, `bearer`, `jwt`, `oauth2`, `oidc` |
| `username` | string | No | - | Username (for basic auth) |
| `password` | string | No | - | Password (for basic auth) |
| `token` | string | No | - | Bearer token |
| `secret` | string | No | - | JWT secret |
| `provider` | string | No | - | OAuth2 provider |
| `client_id` | string | No | - | OAuth2 client ID |
| `client_secret` | string | No | - | OAuth2 client secret |
| `redirect_url` | string | No | - | OAuth2 redirect URL |
| `scopes` | array | No | [] | OAuth2 scopes |

## Path Rewrite Rules

Path rewrite rules use the format `pattern:replacement`:

- `^/api/(.*):/$1` - Remove `/api` prefix
- `^/v1/user/(.*):/user/$1` - Transform path
- `^/old/(.*):/new/$1` - Replace path segment

Patterns use regular expressions. The replacement can reference capture groups.

## Examples

See [Examples](/guide/examples) for complete configuration examples.

## Next Steps

- [Routing](/guide/routing) - Learn about routing and path matching
- [Health Check](/guide/health-check) - Configure health checks
- [Plugins](/guide/plugins) - Extend functionality with plugins
