# API 参考

API Gateway 配置和编程使用的完整 API 参考。

## 配置 API

- [配置](/zh/api/config) - 主配置结构
- [路由](/zh/api/route) - 路由配置
- [插件](/zh/api/plugin) - 插件接口

## 编程使用

API Gateway 可以在 Go 应用程序中以编程方式使用：

```go
import (
    "github.com/go-zoox/api-gateway/core"
    "github.com/go-zoox/api-gateway/config"
)

cfg := &config.Config{
    Port: 8080,
    Routes: []route.Route{
        {
            Name: "api",
            Path: "/api",
            Backend: route.Backend{
                Service: service.Service{
                    Name: "api.example.com",
                    Port: 443,
                    Protocol: "https",
                },
            },
        },
    },
}

app, err := core.New("1.4.5", cfg)
if err != nil {
    // 处理错误
}

err = app.Run()
```

## 类型定义

所有配置类型在以下包中定义：

- `github.com/go-zoox/api-gateway/config` - 主配置
- `github.com/go-zoox/api-gateway/core/route` - 路由定义
- `github.com/go-zoox/api-gateway/core/service` - 服务定义
- `github.com/go-zoox/api-gateway/plugin` - 插件接口

## 另请参阅

- [配置指南](/zh/guide/configuration) - 配置指南
- [示例](/zh/guide/examples) - 配置示例
