# CORS 插件

包路径：`github.com/go-zoox/api-gateway/plugin/cors`

当 **全局** `cors.enable` 为真，**或** **任一路由** 启用 `cors.enable` 时注册。

1. **预检**：`OPTIONS` 且带 `Origin` 与 `Access-Control-Request-Method` 时，由网关直接返回 **204**，并带上 CORS 头，**不转发**上游。
2. **普通请求**：在 **`OnResponse`** 中给上游响应补充 `Access-Control-Allow-Origin` 等头。

若使用 `allow_origins: ["*"]`，则 **不能** 同时开启 `allow_credentials`（`Prepare` 阶段会报错）。

字段与路由覆盖说明见英文版 [CORS](/guide/plugins/cors)。

完整示例见仓库内 `docs/examples/cors.yaml`。

## 另请参阅

- [IP 策略](./ip-policy)
- [配置说明](/zh/guide/configuration)
