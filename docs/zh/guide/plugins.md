# 插件系统

API Gateway 提供插件系统用于扩展功能。

## 插件接口

插件实现 `Plugin` 接口：

```go
type Plugin interface {
    // Prepare 在网关初始化期间调用
    Prepare(app *zoox.Application, cfg *config.Config) error
    
    // OnRequest 在转发请求之前调用
    OnRequest(ctx *zoox.Context, req *http.Request) error
    
    // OnResponse 在接收响应之后调用
    OnResponse(ctx *zoox.Context, res *http.Response) error
}
```

## 内置插件

### BaseURI 插件

BaseURI 插件通过基础 URI 前缀过滤请求。当配置 `baseuri` 时自动启用：

```yaml
baseuri: /v1
```

这确保所有请求必须以 `/v1` 开头。不匹配的请求返回 404 错误。

## 自定义插件

要创建自定义插件：

1. 实现 `Plugin` 接口
2. 在网关注册插件

### 示例插件

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
    // 您的插件配置
}

func (p *MyPlugin) Prepare(app *zoox.Application, cfg *config.Config) error {
    // 初始化您的插件
    // 添加中间件、路由等
    app.Use(func(ctx *zoox.Context) {
        // 您的中间件逻辑
        ctx.Next()
    })
    return nil
}

func (p *MyPlugin) OnRequest(ctx *zoox.Context, req *http.Request) error {
    // 在转发之前修改请求
    req.Header.Set("X-Custom-Header", "value")
    return nil
}

func (p *MyPlugin) OnResponse(ctx *zoox.Context, res *http.Response) error {
    // 在返回之前修改响应
    res.Header.Set("X-Response-Header", "value")
    return nil
}
```

### 注册插件

插件在创建网关时以编程方式注册：

```go
import (
    "github.com/go-zoox/api-gateway/core"
    "github.com/go-zoox/api-gateway/config"
    "myplugin"
)

cfg := &config.Config{
    // ... 配置
}

app, err := core.New(version, cfg)
if err != nil {
    // 处理错误
}

app.Plugin(&myplugin.MyPlugin{
    // 插件配置
})

err = app.Run()
```

## 插件生命周期

1. **Prepare**：在网关初始化期间调用。用于：
   - 注册中间件
   - 设置路由
   - 初始化资源

2. **OnRequest**：在每个请求转发之前调用。用于：
   - 修改请求头
   - 转换请求体
   - 添加身份验证
   - 记录请求

3. **OnResponse**：在每个响应接收之后调用。用于：
   - 修改响应头
   - 转换响应体
   - 添加缓存头
   - 记录响应

## 插件示例

### 身份验证插件

```go
func (p *AuthPlugin) OnRequest(ctx *zoox.Context, req *http.Request) error {
    token := req.Header.Get("Authorization")
    if token == "" {
        return fmt.Errorf("unauthorized")
    }
    
    // 验证 token
    if !p.validateToken(token) {
        return fmt.Errorf("invalid token")
    }
    
    return nil
}
```

### 日志插件

```go
func (p *LoggingPlugin) OnRequest(ctx *zoox.Context, req *http.Request) error {
    p.logger.Infof("请求: %s %s", req.Method, req.URL.Path)
    return nil
}

func (p *LoggingPlugin) OnResponse(ctx *zoox.Context, res *http.Response) error {
    p.logger.Infof("响应: %d", res.StatusCode)
    return nil
}
```

### 限流插件

```go
func (p *RateLimitPlugin) OnRequest(ctx *zoox.Context, req *http.Request) error {
    key := p.getClientKey(req)
    if !p.rateLimiter.Allow(key) {
        return fmt.Errorf("rate limit exceeded")
    }
    return nil
}
```

## 最佳实践

1. **错误处理**：从 `OnRequest` 或 `OnResponse` 返回错误以停止处理
2. **性能**：保持插件逻辑轻量级，避免影响网关性能
3. **线程安全**：如果插件维护状态，确保它们是线程安全的
4. **配置**：使用 config 参数访问网关配置
5. **日志记录**：使用上下文记录器进行一致的日志记录

## 下一步

- [配置](/zh/guide/configuration) - 了解网关配置
- [示例](/zh/guide/examples) - 查看插件示例
- [API 参考](/zh/api/plugin) - 完整的插件 API 文档
