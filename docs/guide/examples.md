# Examples

This page contains practical examples of API Gateway configurations.

## Basic Proxy

Simple reverse proxy configuration:

```yaml
version: v1
port: 8080

routes:
  - name: httpbin
    path: /
    backend:
      service:
        protocol: https
        name: httpbin.org
        port: 443
```

Access: `http://localhost:8080/get` → `https://httpbin.org/get`

## Path Prefix Removal

Remove the path prefix when forwarding:

```yaml
version: v1
port: 8080

routes:
  - name: api
    path: /api
    backend:
      service:
        protocol: https
        name: api.example.com
        port: 443
```

Request: `GET /api/users` → Backend: `GET /users`

## Keep Full Path

Keep the full path when forwarding:

```yaml
version: v1
port: 8080

routes:
  - name: api
    path: /api
    backend:
      service:
        protocol: https
        name: api.example.com
        port: 443
        request:
          path:
            disable_prefix_rewrite: true
```

Request: `GET /api/users` → Backend: `GET /api/users`

## Path Rewriting

Custom path rewriting:

```yaml
version: v1
port: 8080

routes:
  - name: api
    path: /v1
    backend:
      service:
        protocol: https
        name: api.example.com
        port: 443
        request:
          path:
            rewrites:
              - "^/v1/user/(.*):/user/v1/$1"
```

Request: `GET /v1/user/123` → Backend: `GET /user/v1/123`

## Multiple Routes

Route different paths to different backends:

```yaml
version: v1
port: 8080

routes:
  - name: user-service
    path: /api/users
    backend:
      service:
        protocol: https
        name: user-service.example.com
        port: 443
  
  - name: product-service
    path: /api/products
    backend:
      service:
        protocol: https
        name: product-service.example.com
        port: 443
  
  - name: order-service
    path: /api/orders
    backend:
      service:
        protocol: https
        name: order-service.example.com
        port: 443
```

## Regex Routing

Use regex for complex path matching:

```yaml
version: v1
port: 8080

routes:
  - name: user-by-id
    path: ^/api/user/(\d+)
    path_type: regex
    backend:
      service:
        protocol: https
        name: user-service.example.com
        port: 443
        request:
          path:
            rewrites:
              - "^/api/user/(.*):/user/$1"
```

## Base URI Plugin

Use base URI to restrict all routes:

```yaml
version: v1
port: 8080
baseuri: /v1

routes:
  - name: api
    path: /api
    backend:
      service:
        protocol: https
        name: api.example.com
        port: 443
```

Only requests starting with `/v1` are accepted:
- `/v1/api/users` ✓
- `/api/users` ✗ (404)

## Request Headers

Add custom request headers:

```yaml
version: v1
port: 8080

routes:
  - name: api
    path: /api
    backend:
      service:
        protocol: https
        name: api.example.com
        port: 443
        request:
          headers:
            X-API-Key: secret-key
            X-Forwarded-By: api-gateway
```

## Response Headers

Add custom response headers:

```yaml
version: v1
port: 8080

routes:
  - name: api
    path: /api
    backend:
      service:
        protocol: https
        name: api.example.com
        port: 443
        response:
          headers:
            X-Powered-By: api-gateway
            Access-Control-Allow-Origin: "*"
```

## Query Parameters

Add query parameters:

```yaml
version: v1
port: 8080

routes:
  - name: api
    path: /api
    backend:
      service:
        protocol: https
        name: api.example.com
        port: 443
        request:
          query:
            api_key: secret-key
            version: v1
```

## Authentication

### Bearer Token

```yaml
version: v1
port: 8080

routes:
  - name: api
    path: /api
    backend:
      service:
        protocol: https
        name: api.example.com
        port: 443
        auth:
          type: bearer
          token: your-bearer-token
```

### Basic Auth

```yaml
version: v1
port: 8080

routes:
  - name: api
    path: /api
    backend:
      service:
        protocol: https
        name: api.example.com
        port: 443
        auth:
          type: basic
          username: user
          password: pass
```

## Health Check

Enable health checks:

```yaml
version: v1
port: 8080

healthcheck:
  outer:
    enable: true
    path: /healthz
  inner:
    enable: true
    interval: 30
    timeout: 5

routes:
  - name: api
    path: /api
    backend:
      service:
        protocol: https
        name: api.example.com
        port: 443
        health_check:
          enable: true
          path: /health
          status: [200]
```

## Redis Cache

Enable Redis caching:

```yaml
version: v1
port: 8080

cache:
  engine: redis
  host: 127.0.0.1
  port: 6379
  password: ""
  db: 0
  prefix: "api-gateway:"

routes:
  - name: api
    path: /api
    backend:
      service:
        protocol: https
        name: api.example.com
        port: 443
```

## Complete Example

Production-ready configuration:

```yaml
version: v1
port: 8080
baseuri: /v1

cache:
  engine: redis
  host: redis.example.com
  port: 6379
  password: redis-password
  db: 0

healthcheck:
  outer:
    enable: true
    path: /healthz
  inner:
    enable: true
    interval: 30
    timeout: 5

routes:
  - name: user-service
    path: /users
    backend:
      service:
        protocol: https
        name: user-service.example.com
        port: 443
        request:
          path:
            disable_prefix_rewrite: false
          headers:
            X-Service: user-service
        health_check:
          enable: true
          path: /health
          status: [200]
  
  - name: product-service
    path: /products
    backend:
      service:
        protocol: https
        name: product-service.example.com
        port: 443
        request:
          path:
            rewrites:
              - "^/products/(.*):/api/products/$1"
        health_check:
          enable: true
          path: /health
          status: [200]
```

## Docker Compose

Complete setup with Docker:

```yaml
# docker-compose.yml
version: '3.8'

services:
  api-gateway:
    image: gozoox/api-gateway:latest
    ports:
      - "8080:8080"
    volumes:
      - ./config.yaml:/etc/api-gateway/config.yaml
    depends_on:
      - redis
  
  redis:
    image: redis:alpine
    ports:
      - "6379:6379"
```

## Next Steps

- [Configuration](/guide/configuration) - Learn about all configuration options
- [Routing](/guide/routing) - Advanced routing patterns
- [Plugins](/guide/plugins) - Extend functionality with plugins
