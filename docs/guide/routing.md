# Routing

API Gateway provides flexible routing capabilities with support for prefix and regex-based path matching.

## Path Matching

API Gateway supports two types of path matching:

### Prefix Matching

Prefix matching (default) matches requests that start with the specified path:

```yaml
routes:
  - name: api
    path: /api
    path_type: prefix
    backend:
      service:
        name: api.example.com
        port: 8080
```

This will match:
- `/api/users` ✓
- `/api/users/123` ✓
- `/api/v1/data` ✓
- `/apix` ✗ (doesn't start with `/api`)

### Regex Matching

Regex matching uses regular expressions for more complex patterns:

```yaml
routes:
  - name: user-api
    path: ^/v1/user/(\d+)
    path_type: regex
    backend:
      service:
        name: user-api.example.com
        port: 8080
```

This will match:
- `/v1/user/123` ✓
- `/v1/user/456` ✓
- `/v1/user/abc` ✗ (not a number)

## Path Rewriting

Path rewriting allows you to transform the request path before forwarding to the backend.

### Automatic Prefix Removal

By default, when using prefix matching, the matched prefix is automatically removed:

```yaml
routes:
  - name: api
    path: /api
    backend:
      service:
        name: backend.example.com
        port: 8080
```

Request: `GET /api/users` → Backend: `GET /users`

### Disable Prefix Rewrite

To keep the full path:

```yaml
routes:
  - name: api
    path: /api
    backend:
      service:
        name: backend.example.com
        port: 8080
        request:
          path:
            disable_prefix_rewrite: true
```

Request: `GET /api/users` → Backend: `GET /api/users`

### Custom Rewrite Rules

Use rewrite rules for complex transformations:

```yaml
routes:
  - name: api
    path: /api
    backend:
      service:
        name: backend.example.com
        port: 8080
        request:
          path:
            rewrites:
              - "^/api/v1/(.*):/v1/$1"
              - "^/api/v2/(.*):/v2/$1"
```

Rewrite rules use the format `pattern:replacement`:
- Pattern: Regular expression to match
- Replacement: Replacement string (can use capture groups like `$1`, `$2`)

### Rewrite Examples

**Remove prefix:**
```yaml
rewrites:
  - "^/api/(.*):/$1"
```
`/api/users/123` → `/users/123`

**Add prefix:**
```yaml
rewrites:
  - "^/(.*):/api/$1"
```
`/users/123` → `/api/users/123`

**Transform path:**
```yaml
rewrites:
  - "^/v1/user/(.*):/user/v1/$1"
```
`/v1/user/123` → `/user/v1/123`

**Multiple rewrites:**
```yaml
rewrites:
  - "^/old/(.*):/new/$1"
  - "^/new/(.*):/api/$1"
```
Applied in order: `/old/data` → `/new/data` → `/api/data`

## Route Priority

Routes are matched in the order they appear in the configuration. The first matching route is used.

```yaml
routes:
  - name: specific
    path: /api/v1/users
    backend:
      service:
        name: user-service.example.com
        port: 8080
  
  - name: general
    path: /api
    backend:
      service:
        name: api.example.com
        port: 8080
```

In this example, `/api/v1/users` matches the first route, while `/api/v1/posts` matches the second.

## Service Discovery

API Gateway uses DNS for service discovery. The `name` field in service configuration is resolved to an IP address:

```yaml
backend:
  service:
    name: api.example.com  # Resolved via DNS
    port: 443
    protocol: https
```

DNS resolution happens at request time, allowing for dynamic service discovery.

## Request/Response Headers

Add custom headers to requests and responses:

```yaml
routes:
  - name: api
    path: /api
    backend:
      service:
        name: api.example.com
        port: 8080
        request:
          headers:
            X-Forwarded-For: gateway
            X-Custom-Header: value
        response:
          headers:
            X-Powered-By: api-gateway
            Access-Control-Allow-Origin: "*"
```

## Query Parameters

Add query parameters to requests:

```yaml
routes:
  - name: api
    path: /api
    backend:
      service:
        name: api.example.com
        port: 8080
        request:
          query:
            api_key: secret-key
            version: v1
```

## Examples

See [Examples](/guide/examples) for more routing examples.

## Next Steps

- [Configuration](/guide/configuration) - Complete configuration reference
- [Health Check](/guide/health-check) - Configure health checks for routes
- [Plugins](/guide/plugins) - Extend routing with plugins
