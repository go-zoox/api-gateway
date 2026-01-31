# 使用示例

本页包含 API Gateway 配置的实用示例。

## 基础代理

简单的反向代理配置：

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

访问：`http://localhost:8080/get` → `https://httpbin.org/get`

## 路径前缀移除

转发时移除路径前缀：

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

请求：`GET /api/users` → 后端：`GET /users`

## 保留完整路径

转发时保留完整路径：

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

请求：`GET /api/users` → 后端：`GET /api/users`

## 路径重写

自定义路径重写：

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

请求：`GET /v1/user/123` → 后端：`GET /user/v1/123`

## 多路由

将不同路径路由到不同后端：

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

## 正则路由

使用正则表达式进行复杂的路径匹配：

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

## BaseURI 插件

使用基础 URI 限制所有路由：

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

只接受以 `/v1` 开头的请求：
- `/v1/api/users` ✓
- `/api/users` ✗ (404)

## 请求头

添加自定义请求头：

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

## 响应头

添加自定义响应头：

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

## 查询参数

添加查询参数：

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

## 身份验证

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

### 基本认证

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

## 健康检查

启用健康检查：

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

## Redis 缓存

启用 Redis 缓存：

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

## 完整示例

生产就绪配置：

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

使用 Docker 的完整设置：

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

## 下一步

- [配置](/zh/guide/configuration) - 了解所有配置选项
- [路由](/zh/guide/routing) - 高级路由模式
- [插件](/zh/guide/plugins) - 使用插件扩展功能
