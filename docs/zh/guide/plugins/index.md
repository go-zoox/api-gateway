# 插件系统

API Gateway 将**插件**单独成篇：`github.com/go-zoox/api-gateway/plugin/…` 下每个子包对应一类能力。本节说明统一的 `Plugin` 接口、请求生命周期，并通过**每个插件一篇文档**介绍内置插件。

## 内置插件

| 插件 | 包路径 | 何时启用 |
| --- | --- | --- |
| [Base URI](./base-uri) | `plugin/baseuri` | 配置项 `baseuri` 非空 |
| [限流](./rate-limit) | `plugin/ratelimit` | 全局 `rate_limit.enable` 为真，或任一路由启用 `rate_limit.enable` |

## 插件接口

```go
type Plugin interface {
    Prepare(app *zoox.Application, cfg *config.Config) error
    OnRequest(ctx *zoox.Context, req *http.Request) error
    OnResponse(ctx *zoox.Context, res *http.Response) error
}
```

`Prepare` 在启动时执行一次；`OnRequest` 在转发到上游之前调用；`OnResponse` 在收到上游响应之后调用。

## 生命周期

1. **Prepare**：注册中间件、建立连接、读取配置。
2. **OnRequest**：鉴权、改写、限流、日志等。
3. **OnResponse**：修改响应头/体、采集指标等。

若 `OnRequest` 或 `OnResponse` 返回非 nil 错误，通常会终止该请求后续处理并向客户端返回 HTTP 错误。

## 自定义插件

1. 嵌入 `plugin.Plugin`（可选），并实现上述三个方法。
2. 构建网关实例时挂载插件：

```go
import (
    "github.com/go-zoox/api-gateway/config"
    "github.com/go-zoox/api-gateway/core"
    "github.com/go-zoox/api-gateway/plugin"
)

app, err := core.New(version, cfg)
if err != nil {
    return err
}

app.Plugin(&MyPlugin{})
return app.Run()
```

### 示例骨架

```go
type MyPlugin struct {
    plugin.Plugin
}

func (p *MyPlugin) Prepare(app *zoox.Application, cfg *config.Config) error {
    app.Use(func(ctx *zoox.Context) {
        ctx.Next()
    })
    return nil
}

func (p *MyPlugin) OnRequest(ctx *zoox.Context, req *http.Request) error {
    req.Header.Set("X-Example", "1")
    return nil
}

func (p *MyPlugin) OnResponse(ctx *zoox.Context, res *http.Response) error {
    return nil
}
```

## 最佳实践

1. `OnRequest` / `OnResponse` 保持轻量；重任务放到异步队列或 worker。
2. 若插件持有可变共享状态，必须保证并发安全。
3. 使用 `ctx.Logger` 输出日志以保持一致格式。
4. 优先通过 `cfg` 与路由结构体做类型化配置。

## 延伸阅读

- [Base URI](./base-uri) — 统一 URL 前缀并剥离路径。
- [限流](./rate-limit) — 令牌桶、漏桶、固定窗口；内存或 Redis。
- [配置说明](/zh/guide/configuration) — YAML 全局项与路由。
- [插件 API](/zh/api/plugin) — `plugin` 包说明。
