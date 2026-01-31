# 插件 API

`Plugin` 接口允许扩展 API Gateway 功能。

## 接口定义

```go
type Plugin interface {
    Prepare(app *zoox.Application, cfg *config.Config) error
    OnRequest(ctx *zoox.Context, req *http.Request) error
    OnResponse(ctx *zoox.Context, res *http.Response) error
}
```

## 方法

### Prepare

```go
Prepare(app *zoox.Application, cfg *config.Config) error
```

在网关初始化期间调用。用于：
- 注册中间件
- 设置路由
- 初始化资源

**参数：**
- `app` (*zoox.Application): Zoox 应用程序实例
- `cfg` (*config.Config): 网关配置

**返回：**
- `error`: 如果初始化失败则返回错误

### OnRequest

```go
OnRequest(ctx *zoox.Context, req *http.Request) error
```

在将请求转发到后端之前调用。用于：
- 修改请求头
- 转换请求体
- 添加身份验证
- 记录请求

**参数：**
- `ctx` (*zoox.Context): 请求上下文
- `req` (*http.Request): HTTP 请求（可以修改）

**返回：**
- `error`: 错误以停止处理并返回错误响应

### OnResponse

```go
OnResponse(ctx *zoox.Context, res *http.Response) error
```

在从后端接收响应后调用。用于：
- 修改响应头
- 转换响应体
- 添加缓存头
- 记录响应

**参数：**
- `ctx` (*zoox.Context): 请求上下文
- `res` (*http.Response): HTTP 响应（可以修改）

**返回：**
- `error`: 错误以返回错误响应

## 示例实现

```go
package myplugin

import (
    "net/http"
    "github.com/go-zoox/api-gateway/config"
    "github.com/go-zoox/api-gateway/plugin"
    "github.com/go-zoox/zoox"
)

type MyPlugin struct {
    plugin.Plugin
    // 插件配置
}

func (p *MyPlugin) Prepare(app *zoox.Application, cfg *config.Config) error {
    // 初始化插件
    app.Use(func(ctx *zoox.Context) {
        // 中间件逻辑
        ctx.Next()
    })
    return nil
}

func (p *MyPlugin) OnRequest(ctx *zoox.Context, req *http.Request) error {
    // 修改请求
    req.Header.Set("X-Custom-Header", "value")
    return nil
}

func (p *MyPlugin) OnResponse(ctx *zoox.Context, res *http.Response) error {
    // 修改响应
    res.Header.Set("X-Response-Header", "value")
    return nil
}
```

## 注册插件

插件以编程方式注册：

```go
app, err := core.New(version, cfg)
if err != nil {
    // 处理错误
}

app.Plugin(&myplugin.MyPlugin{
    // 配置
})

err = app.Run()
```

## 内置插件

### BaseURI 插件

当配置 `baseuri` 时，BaseURI 插件会自动启用：

```go
type BaseURI struct {
    plugin.Plugin
    BaseURI string
}
```

此插件通过基础 URI 前缀过滤请求。

## 最佳实践

1. **错误处理**：返回错误以停止处理
2. **性能**：保持插件逻辑轻量级
3. **线程安全**：确保插件是线程安全的
4. **日志记录**：使用上下文记录器进行一致的日志记录
5. **配置**：通过 `cfg` 参数访问网关配置

## 另请参阅

- [插件指南](/zh/guide/plugins) - 插件开发指南
- [配置 API](/zh/api/config) - 配置 API
