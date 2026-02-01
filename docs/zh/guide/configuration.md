# 配置说明

API Gateway 使用 YAML 配置文件。本文档描述了所有可用的配置选项。

## 配置文件结构

```yaml
version: v1              # 配置版本
port: 8080               # 网关监听端口
baseuri: /v1             # 基础 URI 前缀（可选）

cache:                   # 缓存配置（可选）
  engine: redis
  host: 127.0.0.1
  port: 6379
  password: ""
  db: 0

healthcheck:             # 健康检查配置
  outer:                 # 外部健康检查
    enable: true
    path: /healthz
    ok: true
  inner:                 # 内部服务健康检查
    enable: true
    interval: 30         # 检查间隔（秒）
    timeout: 5           # 超时时间（秒）

backend:                 # 默认后端（可选）
  service:
    protocol: https
    name: example.com
    port: 443

routes:                  # 路由定义
  - name: route-name
    path: /api
    path_type: prefix    # prefix 或 regex
    backend:
      service:
        protocol: https
        name: backend.example.com
        port: 443
        request:
          path:
            disable_prefix_rewrite: false
            rewrites:
              - "^/api/(.*):/$1"
          headers:
            X-Custom-Header: value
          query:
            key: value
        response:
          headers:
            X-Response-Header: value
        auth:
          type: bearer
          token: your-token
```

## 配置字段

### 顶级字段

| 字段 | 类型 | 必需 | 默认值 | 描述 |
|------|------|------|--------|------|
| `version` | string | 是 | - | 配置版本（当前为 `v1`） |
| `port` | int | 否 | 8080 | 网关监听的端口 |
| `baseuri` | string | 否 | - | 所有路由的基础 URI 前缀 |
| `cache` | object | 否 | - | 缓存配置 |
| `healthcheck` | object | 否 | - | 健康检查配置 |
| `backend` | object | 否 | - | 默认后端服务 |
| `routes` | array | 否 | [] | 路由定义 |

### 缓存配置

| 字段 | 类型 | 必需 | 默认值 | 描述 |
|------|------|------|--------|------|
| `engine` | string | 否 | redis | 缓存引擎（当前仅支持 `redis`） |
| `host` | string | 否 | 127.0.0.1 | Redis 主机 |
| `port` | int | 否 | 6379 | Redis 端口 |
| `password` | string | 否 | - | Redis 密码 |
| `db` | int | 否 | 0 | Redis 数据库编号 |
| `prefix` | string | 否 | gozoox-api-gateway: | 键前缀 |

### 健康检查配置

#### 外部健康检查

| 字段 | 类型 | 必需 | 默认值 | 描述 |
|------|------|------|--------|------|
| `enable` | bool | 否 | false | 启用外部健康检查端点 |
| `path` | string | 否 | /healthz | 健康检查端点路径 |
| `ok` | bool | 否 | true | 始终返回 OK（跳过实际检查） |

#### 内部健康检查

| 字段 | 类型 | 必需 | 默认值 | 描述 |
|------|------|------|--------|------|
| `enable` | bool | 否 | false | 启用内部服务健康检查 |
| `interval` | int | 否 | 30 | 检查间隔（秒） |
| `timeout` | int | 否 | 5 | 请求超时（秒） |

### 路由配置

| 字段 | 类型 | 必需 | 默认值 | 描述 |
|------|------|------|--------|------|
| `name` | string | 是 | - | 路由名称（用于日志记录） |
| `path` | string | 是 | - | 要匹配的路径模式 |
| `path_type` | string | 否 | prefix | 匹配类型：`prefix` 或 `regex` |
| `backend` | object | 是 | - | 后端服务配置 |

### 服务配置

| 字段 | 类型 | 必需 | 默认值 | 描述 |
|------|------|------|--------|------|
| `protocol` | string | 否 | http | 协议：`http` 或 `https` |
| `name` | string | 否* | - | 服务主机名或 IP（单服务模式必需） |
| `port` | int | 否* | 80 | 服务端口（单服务模式必需） |
| `algorithm` | string | 否 | round-robin | 负载均衡算法：`round-robin`、`weighted`、`least-connections`、`ip-hash` |
| `servers` | array | 否 | [] | 负载均衡的服务器实例列表（如果设置，启用多服务模式） |
| `request` | object | 否 | - | 请求转换 |
| `response` | object | 否 | - | 响应转换 |
| `auth` | object | 否 | - | 身份验证配置 |
| `health_check` | object | 否 | - | 服务特定的健康检查 |

**注意：** 单服务模式下，`name` 和 `port` 是必需的。多服务模式下，`servers` 数组是必需的。

#### 服务器配置（多服务模式）

| 字段 | 类型 | 必需 | 默认值 | 描述 |
|------|------|------|--------|------|
| `name` | string | 是 | - | 服务器主机名或 IP |
| `port` | int | 是 | - | 服务器端口 |
| `protocol` | string | 否 | - | 协议（未设置时继承服务配置） |
| `weight` | int | 否 | 1 | 权重（用于 weighted 算法） |
| `disabled` | bool | 否 | false | 禁用服务器（服务器默认启用） |
| `request` | object | 否 | - | 服务器特定的请求配置（覆盖全局） |
| `response` | object | 否 | - | 服务器特定的响应配置（覆盖全局） |
| `auth` | object | 否 | - | 服务器特定的身份验证（覆盖全局） |
| `health_check` | object | 否 | - | 服务器特定的健康检查（覆盖全局） |

### 请求配置

| 字段 | 类型 | 必需 | 默认值 | 描述 |
|------|------|------|--------|------|
| `path` | object | 否 | - | 路径重写配置 |
| `headers` | map | 否 | - | 额外的请求头 |
| `query` | map | 否 | - | 额外的查询参数 |

#### 路径重写

| 字段 | 类型 | 必需 | 默认值 | 描述 |
|------|------|------|--------|------|
| `disable_prefix_rewrite` | bool | 否 | false | 禁用自动前缀移除 |
| `rewrites` | array | 否 | [] | 路径重写规则（格式：`pattern:replacement`） |

### 响应配置

| 字段 | 类型 | 必需 | 默认值 | 描述 |
|------|------|------|--------|------|
| `headers` | map | 否 | - | 额外的响应头 |

### 身份验证配置

| 字段 | 类型 | 必需 | 默认值 | 描述 |
|------|------|------|--------|------|
| `type` | string | 否 | - | 认证类型：`basic`、`bearer`、`jwt`、`oauth2`、`oidc` |
| `username` | string | 否 | - | 用户名（用于基本认证） |
| `password` | string | 否 | - | 密码（用于基本认证） |
| `token` | string | 否 | - | Bearer token |
| `secret` | string | 否 | - | JWT 密钥 |
| `provider` | string | 否 | - | OAuth2 提供商 |
| `client_id` | string | 否 | - | OAuth2 客户端 ID |
| `client_secret` | string | 否 | - | OAuth2 客户端密钥 |
| `redirect_url` | string | 否 | - | OAuth2 重定向 URL |
| `scopes` | array | 否 | [] | OAuth2 作用域 |

## 路径重写规则

路径重写规则使用格式 `pattern:replacement`：

- `^/api/(.*):/$1` - 移除 `/api` 前缀
- `^/v1/user/(.*):/user/$1` - 转换路径
- `^/old/(.*):/new/$1` - 替换路径段

模式使用正则表达式。替换可以引用捕获组。

## 示例

查看[示例](/zh/guide/examples)了解完整的配置示例。

## 负载均衡

有关负载均衡配置，请参阅[负载均衡](/zh/guide/load-balancing)指南。

## 下一步

- [路由](/zh/guide/routing) - 了解路由和路径匹配
- [负载均衡](/zh/guide/load-balancing) - 配置负载均衡
- [健康检查](/zh/guide/health-check) - 配置健康检查
- [插件](/zh/guide/plugins) - 使用插件扩展功能
