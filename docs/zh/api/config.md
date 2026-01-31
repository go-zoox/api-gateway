# 配置 API

`Config` 结构体定义了 API Gateway 的主配置。

## 类型定义

```go
type Config struct {
    Port     int64          `config:"port"`
    BaseURI  string         `config:"baseuri"`
    Backend  route.Backend  `config:"backend"`
    Routes   []route.Route  `config:"routes"`
    Cache    Cache          `config:"cache"`
    HealthCheck HealthCheck `config:"healthcheck"`
}
```

## 字段

### Port

```go
Port int64
```

网关监听的端口。默认值：`8080`

### BaseURI

```go
BaseURI string
```

所有路由的基础 URI 前缀。设置后，只接受以此前缀开头的请求。

### Backend

```go
Backend route.Backend
```

默认后端服务。当没有路由匹配时使用。

### Routes

```go
Routes []route.Route
```

路由定义数组。有关详细信息，请参阅[路由 API](/zh/api/route)。

### Cache

```go
Cache Cache
```

缓存配置。请参阅[缓存配置](#缓存配置)。

### HealthCheck

```go
HealthCheck HealthCheck
```

健康检查配置。请参阅[健康检查配置](#健康检查配置)。

## 缓存配置

```go
type Cache struct {
    Host     string `config:"host"`
    Port     int64  `config:"port"`
    Username string `config:"username"`
    Password string `config:"password"`
    DB       int64  `config:"db"`
    Prefix   string `config:"prefix"`
}
```

### 字段

- `Host` (string): Redis 主机。默认值：`127.0.0.1`
- `Port` (int64): Redis 端口。默认值：`6379`
- `Username` (string): Redis 用户名（可选）
- `Password` (string): Redis 密码（可选）
- `DB` (int64): Redis 数据库编号。默认值：`0`
- `Prefix` (string): 键前缀。默认值：`gozoox-api-gateway:`

## 健康检查配置

```go
type HealthCheck struct {
    Outer HealthCheckOuter `config:"outer"`
    Inner HealthCheckInner `config:"inner"`
}
```

### 外部健康检查

```go
type HealthCheckOuter struct {
    Enable bool   `config:"enable"`
    Path   string `config:"path"`
    Ok     bool   `config:"ok"`
}
```

- `Enable` (bool): 启用外部健康检查端点
- `Path` (string): 健康检查端点路径。默认值：`/healthz`
- `Ok` (bool): 始终返回 OK。默认值：`true`

### 内部健康检查

```go
type HealthCheckInner struct {
    Enable   bool `config:"enable"`
    Interval int64 `config:"interval"`
    Timeout  int64 `config:"timeout"`
}
```

- `Enable` (bool): 启用内部服务健康检查
- `Interval` (int64): 检查间隔（秒）。默认值：`30`
- `Timeout` (int64): 请求超时（秒）。默认值：`5`

## 示例

```go
cfg := &config.Config{
    Port: 8080,
    BaseURI: "/v1",
    Cache: config.Cache{
        Host: "127.0.0.1",
        Port: 6379,
        DB: 0,
    },
    HealthCheck: config.HealthCheck{
        Outer: config.HealthCheckOuter{
            Enable: true,
            Path: "/healthz",
        },
        Inner: config.HealthCheckInner{
            Enable: true,
            Interval: 30,
            Timeout: 5,
        },
    },
    Routes: []route.Route{
        // ... 路由
    },
}
```

## 另请参阅

- [路由 API](/zh/api/route) - 路由配置
- [配置指南](/zh/guide/configuration) - 配置指南
