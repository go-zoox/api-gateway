# 路由 API

`Route` 结构体定义了 API Gateway 的路由配置。

## 类型定义

```go
type Route struct {
    Name     string  `config:"name"`
    Path     string  `config:"path"`
    Backend  Backend `config:"backend"`
    PathType string  `config:"path_type,default=prefix"`
}
```

## 字段

### Name

```go
Name string
```

用于日志记录和识别的路由名称。必需。

### Path

```go
Path string
```

要匹配的路径模式。根据 `PathType` 可以是前缀或正则表达式模式。必需。

### Backend

```go
Backend Backend
```

后端服务配置。请参阅[后端配置](#后端配置)。必需。

### PathType

```go
PathType string
```

路径匹配类型。选项：
- `prefix`（默认）：前缀匹配
- `regex`：正则表达式匹配

## 后端配置

```go
type Backend struct {
    Service service.Service `config:"service"`
}
```

### 服务

```go
type Service struct {
    Name        string   `config:"name"`
    Port        int64    `config:"port"`
    Protocol    string   `config:"protocol,default=http"`
    Request     Request  `config:"request"`
    Response    Response `config:"response"`
    Auth        Auth     `config:"auth"`
    HealthCheck HealthCheck `config:"health_check"`
}
```

#### 服务字段

- `Name` (string): 服务主机名或 IP 地址。必需。
- `Port` (int64): 服务端口。默认值：`80`
- `Protocol` (string): 协议（`http` 或 `https`）。默认值：`http`
- `Request` (Request): 请求转换配置
- `Response` (Response): 响应转换配置
- `Auth` (Auth): 身份验证配置
- `HealthCheck` (HealthCheck): 服务特定的健康检查

## 请求配置

```go
type Request struct {
    Path    RequestPath        `config:"path"`
    Headers map[string]string  `config:"headers"`
    Query   map[string]string  `config:"query"`
}
```

### 请求路径

```go
type RequestPath struct {
    DisablePrefixRewrite bool     `config:"disable_prefix_rewrite"`
    Rewrites             []string `config:"rewrites"`
}
```

- `DisablePrefixRewrite` (bool): 禁用自动前缀移除。默认值：`false`
- `Rewrites` ([]string): 路径重写规则，格式为 `pattern:replacement`

### 请求头

```go
Headers map[string]string
```

要添加到请求的额外头。

### 请求查询

```go
Query map[string]string
```

要添加到请求的额外查询参数。

## 响应配置

```go
type Response struct {
    Headers map[string]string `config:"headers"`
}
```

### 响应头

```go
Headers map[string]string
```

要添加到响应的额外头。

## 身份验证配置

```go
type Auth struct {
    Type         string   `config:"type"`
    Username     string   `config:"username"`
    Password     string   `config:"password"`
    Token        string   `config:"token"`
    Secret       string   `config:"secret"`
    Provider     string   `config:"provider"`
    ClientID     string   `config:"client_id"`
    ClientSecret string   `config:"client_secret"`
    RedirectURL  string   `config:"redirect_url"`
    Scopes       []string `config:"scopes"`
}
```

### 认证类型

- `basic`: 基本认证（需要 `username` 和 `password`）
- `bearer`: Bearer token 认证（需要 `token`）
- `jwt`: JWT 认证（需要 `secret`）
- `oauth2`: OAuth2 认证（需要 `provider`、`client_id`、`client_secret` 等）
- `oidc`: OpenID Connect 认证

## 服务健康检查

```go
type HealthCheck struct {
    Enable bool    `config:"enable"`
    Method string  `config:"method,default=GET"`
    Path   string  `config:"path,default=/health"`
    Status []int64 `config:"status,default=[200]"`
    Ok     bool    `config:"ok"`
}
```

- `Enable` (bool): 为此服务启用健康检查
- `Method` (string): HTTP 方法。默认值：`GET`
- `Path` (string): 健康检查端点。默认值：`/health`
- `Status` ([]int64): 有效的 HTTP 状态码。默认值：`[200]`
- `Ok` (bool): 始终认为服务健康。默认值：`false`

## 示例

```go
route := route.Route{
    Name: "api",
    Path: "/api",
    PathType: "prefix",
    Backend: route.Backend{
        Service: service.Service{
            Name: "api.example.com",
            Port: 443,
            Protocol: "https",
            Request: service.Request{
                Path: service.RequestPath{
                    DisablePrefixRewrite: false,
                    Rewrites: []string{
                        "^/api/(.*):/$1",
                    },
                },
                Headers: map[string]string{
                    "X-API-Key": "secret",
                },
            },
            Response: service.Response{
                Headers: map[string]string{
                    "X-Powered-By": "api-gateway",
                },
            },
            HealthCheck: service.HealthCheck{
                Enable: true,
                Path: "/health",
                Status: []int64{200},
            },
        },
    },
}
```

## 另请参阅

- [配置 API](/zh/api/config) - 主配置
- [插件 API](/zh/api/plugin) - 插件接口
- [路由指南](/zh/guide/routing) - 路由指南
