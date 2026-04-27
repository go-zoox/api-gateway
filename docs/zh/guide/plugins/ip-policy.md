# IP 策略插件

包路径：`github.com/go-zoox/api-gateway/plugin/ippolicy`

当 **全局** `ip_policy.enable` 为真，**或** **任一路由** 启用 `ip_policy.enable` 时注册。策略在 **HTTP 中间件**里执行，位于反向代理之前，因此对所有 HTTP 方法都生效（包括 CORS 预检的 `OPTIONS`）。

## 行为说明

- **deny** 中的 CIDR **先**匹配；命中则返回 **403 Forbidden**，响应体为配置的 **`message`**。
- 若 **allow** 非空，客户端 IP 还必须至少命中一条 allow CIDR；若 **allow** 为空，则只按 deny 拦截，其余放行（默认宽松）。
- **trusted_proxies**：若为空，网关**只**使用直连 TCP 对端地址（`RemoteAddr`），**不采用** `X-Forwarded-For` / `X-Real-IP`。若非空，仅当直连对端地址落在这些 CIDR 内时，才把 `X-Forwarded-For` 的**第一段**视为客户端地址（与常见七层负载均衡后的部署方式一致）。

## 配置

### 全局

```yaml
ip_policy:
  enable: true
  allow:
    - 10.0.0.0/8
    - 2001:db8::/32
  deny:
    - 198.51.100.0/24
  trusted_proxies:
    - 10.0.0.0/8
  message: "Forbidden"
```

### 按路由

在路由上设置 `ip_policy.enable`。该路由的 **allow**、**deny**、**trusted_proxies**、**message** 会与**全局**合并：路由上未写的空列表会继承全局；若路由的 **deny** 非空，会在**全局 deny 之后继续追加**；若路由的 **allow** 非空，则在该路由上**以路由的 allow 覆盖**全局的 allow 列表。

## 字段说明

**是否必须** 表示在已经编写 `ip_policy` 块的前提下，该字段是否“必须显式配置才有意义”。**默认** 为省略该字段时生效的值。要让插件真正生效，须在某处将 **`enable` 置为真**（全局或路由），见下表对 `enable` 的说明。

| 字段 | 类型 | 是否必须 | 默认 | 说明 |
| --- | --- | --- | --- | --- |
| `enable` | bool | 否* | `false` | *要注册并运行 IP 策略插件，全局或任一路由上须为 `true`。 |
| `allow` | CIDR 列表 | 否 | 空 | 非空时客户端须命中其中一条；为空时不做 allow 限制（仅看 `deny`）。 |
| `deny` | CIDR 列表 | 否 | 空 | 非空时命中则 403。 |
| `trusted_proxies` | CIDR 列表 | 否 | 空 | 空则只看直连 `RemoteAddr`；非空时仅当对端在集合内才信任 `X-Forwarded-For`（首段）/ `X-Real-IP`。 |
| `message` | 字符串 | 否 | `Forbidden` | 拒绝时响应体。 |

CIDR 可写 `192.0.2.0/24`，也可只写地址（无 `/`），会按单主机处理（IPv4 为 `/32`，IPv6 为 `/128`）。

## 另请参阅

- 仓库内完整示例：`docs/examples/ip-policy.yaml`。
- [CORS](./cors) — 常与安全 IP 范围配合使用。
- [配置说明](/zh/guide/configuration) — 全局与路由结构。
