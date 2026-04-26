# IP 策略插件

包路径：`github.com/go-zoox/api-gateway/plugin/ippolicy`

当 **全局** `ip_policy.enable` 为真，**或** **任一路由** 启用 `ip_policy.enable` 时注册。在 **HTTP 中间件**（反向代理之前）执行，对所有方法生效（含 CORS 预检的 `OPTIONS`）。

## 行为概要

- **deny** 中的 CIDR 优先匹配，命中则返回 **403**，响应体为 **`message`**。
- **allow** 非空时，客户端 IP 还须命中至少一条 allow；**allow** 为空时仅按 deny 过滤（默认宽松）。
- **trusted_proxies** 为空时，**只**看直连对端（`RemoteAddr`），**不使用** `X-Forwarded-For` / `X-Real-IP`。非空时，仅当直连对端落在这些 CIDR 内，才使用 `X-Forwarded-For` 的**第一段**作为客户端 IP。

配置项与合并规则见英文版 [IP policy](/guide/plugins/ip-policy)。

完整示例见仓库内 `docs/examples/ip-policy.yaml`。

## 另请参阅

- [CORS](./cors)
- [配置说明](/zh/guide/configuration)
