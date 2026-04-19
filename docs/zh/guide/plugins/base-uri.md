# Base URI 插件

包路径：`github.com/go-zoox/api-gateway/plugin/baseuri`

当网关配置了全局 Base URI 时，该插件会注册中间件，完成两件事：

1. **拒绝**不以配置前缀开头的路径（返回 **404 Not Found**）。
2. **剥离**该前缀，更新 `ctx.Request.URL.Path` 与 `ctx.Path`，使下游路由无需重复书写前缀。

适用于将 API 暴露在 `/api`、`/v1` 等前缀下，而路由仍按根路径编写的场景。

## 配置

在 YAML 中设置顶层字段 `baseuri`：

```yaml
baseuri: /v1
```

例如 `baseuri: /v1` 时，请求 `GET /v1/users` 在进入路由匹配前会被规范为路径 `/users`。

启动时若 `baseuri` 非空，`core` 会自动挂载该插件（见 `preparePluginsBuildin`）。

## 行为说明

| 请求路径 | 结果 |
| --- | --- |
| 以配置的 base URI 开头 | 去掉前缀后继续处理 |
| 不以该前缀开头 | **404**，不再转发 |

仅整段前缀被剥离一次；其余路径语义仍由网关路由规则决定。

## 与其他插件的关系

Base URI 执行后，路由与其它插件看到的是**剥离后**的路径。启用 base URI 时，`routes` 里的 `path` 建议按剥离后的路径编写。

## 相关链接

- [插件总览](./)
- [路由](/zh/guide/routing)
- [配置](/zh/guide/configuration)
