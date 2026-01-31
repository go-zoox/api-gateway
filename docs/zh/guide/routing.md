# 路由配置

API Gateway 提供灵活的路由功能，支持前缀和基于正则表达式的路径匹配。

## 路径匹配

API Gateway 支持两种类型的路径匹配：

### 前缀匹配

前缀匹配（默认）匹配以指定路径开头的请求：

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

这将匹配：
- `/api/users` ✓
- `/api/users/123` ✓
- `/api/v1/data` ✓
- `/apix` ✗（不以 `/api` 开头）

### 正则匹配

正则匹配使用正则表达式进行更复杂的模式匹配：

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

这将匹配：
- `/v1/user/123` ✓
- `/v1/user/456` ✓
- `/v1/user/abc` ✗（不是数字）

## 路径重写

路径重写允许您在转发到后端之前转换请求路径。

### 自动前缀移除

默认情况下，使用前缀匹配时，匹配的前缀会自动移除：

```yaml
routes:
  - name: api
    path: /api
    backend:
      service:
        name: backend.example.com
        port: 8080
```

请求：`GET /api/users` → 后端：`GET /users`

### 禁用前缀重写

要保留完整路径：

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

请求：`GET /api/users` → 后端：`GET /api/users`

### 自定义重写规则

使用重写规则进行复杂的转换：

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

重写规则使用格式 `pattern:replacement`：
- 模式：要匹配的正则表达式
- 替换：替换字符串（可以使用捕获组，如 `$1`、`$2`）

### 重写示例

**移除前缀：**
```yaml
rewrites:
  - "^/api/(.*):/$1"
```
`/api/users/123` → `/users/123`

**添加前缀：**
```yaml
rewrites:
  - "^/(.*):/api/$1"
```
`/users/123` → `/api/users/123`

**转换路径：**
```yaml
rewrites:
  - "^/v1/user/(.*):/user/v1/$1"
```
`/v1/user/123` → `/user/v1/123`

**多个重写：**
```yaml
rewrites:
  - "^/old/(.*):/new/$1"
  - "^/new/(.*):/api/$1"
```
按顺序应用：`/old/data` → `/new/data` → `/api/data`

## 路由优先级

路由按照配置中出现的顺序进行匹配。使用第一个匹配的路由。

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

在此示例中，`/api/v1/users` 匹配第一个路由，而 `/api/v1/posts` 匹配第二个。

## 服务发现

API Gateway 使用 DNS 进行服务发现。服务配置中的 `name` 字段会解析为 IP 地址：

```yaml
backend:
  service:
    name: api.example.com  # 通过 DNS 解析
    port: 443
    protocol: https
```

DNS 解析在请求时进行，允许动态服务发现。

## 请求/响应头

向请求和响应添加自定义头：

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

## 查询参数

向请求添加查询参数：

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

## 示例

查看[示例](/zh/guide/examples)了解更多路由示例。

## 下一步

- [配置](/zh/guide/configuration) - 完整的配置参考
- [健康检查](/zh/guide/health-check) - 为路由配置健康检查
- [插件](/zh/guide/plugins) - 使用插件扩展路由
