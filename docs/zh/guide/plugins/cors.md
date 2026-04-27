# CORS 插件

包路径：`github.com/go-zoox/api-gateway/plugin/cors`

当 **全局** `cors.enable` 为真，**或** **任一路由** 启用 `cors.enable` 时注册。插件做两件事：

1. **预检请求**：`OPTIONS` 且带 `Origin` 与 `Access-Control-Request-Method` 时，网关直接返回 **204 No Content** 及 CORS 相关响应头，**不转发**到上游。
2. **普通请求**：在收到上游响应后的 **`OnResponse`** 中补充 CORS 头（例如 `Access-Control-Allow-Origin`）。

`allow_origins: ["*"]` 与 **`allow_credentials: true` 不能同时用**（浏览器 CORS 规范要求：带凭据时不能对任意来源使用 `*`，本插件在 `Prepare` 时也会校验并报错）。仅公开、不带 cookie/凭证类跨站访问时，可用 `*` 表示任意来源。

## 配置

### 全局

```yaml
cors:
  enable: true
  allow_origins:
    - "https://app.example.com"
  allow_methods:
    - GET
    - POST
    - PUT
  allow_headers:
    - "*"
  expose_headers:
    - X-Request-Id
  allow_credentials: true
  max_age: 86400
```

### 按路由

在对应路由上设置 `cors.enable` 为真。路由里**未写的数组类字段**会从上面的**全局** `cors` 块继承；需要覆盖的字段在路由上单独写即可。

## 字段说明

**是否必须** 表示在已经编写 `cors` 块的前提下，该字段是否“必须显式配置才有意义”。**默认** 为省略该字段时生效的值。要让插件真正生效，须在某处将 **`enable` 置为真**（全局或路由），见下表对 `enable` 的说明。

| 字段 | 类型 | 是否必须 | 默认 | 说明 |
| --- | --- | --- | --- | --- |
| `enable` | bool | 否* | `false` | *要注册并运行 CORS 插件，全局或任一路由上须为 `true`（与“字段本身”是否必填是两层含义）。 |
| `allow_origins` | 字符串列表 | 否 | `["*"]` | 经规范化后，省略等价于 `*`（任意来源，且不能与 `allow_credentials: true` 同用）。 |
| `allow_methods` | 列表 | 否 | `GET` `POST` `PUT` `PATCH` `DELETE` `HEAD` `OPTIONS` | 预检与 `Access-Control-Allow-Methods`。 |
| `allow_headers` | 列表 | 否 | `["*"]` | 预检响应中的 `Access-Control-Allow-Headers`。 |
| `expose_headers` | 列表 | 否 | 空 | 非预检响应中的 `Access-Control-Expose-Headers`。 |
| `allow_credentials` | bool | 否 | `false` | 为真时 `allow_origins` 中不允许出现 `*`，否则 **Prepare** 报错。 |
| `max_age` | 整数（秒） | 否 | `0` | 预检缓存（`Access-Control-Max-Age`）；`0` 表示不输出该头。 |

## 另请参阅

- 仓库内完整示例：`docs/examples/cors.yaml`。
- [IP 策略](./ip-policy) — 在不可信网络可先收敛来源 IP，再配 CORS。
- [配置说明](/zh/guide/configuration) — 全局与路由结构。
