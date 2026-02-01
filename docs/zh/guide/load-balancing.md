# 负载均衡

API Gateway 支持多种负载均衡算法，可以在多个后端服务实例之间分配流量。

## 概述

负载均衡允许您：
- 在多个后端实例之间分配流量
- 提高可用性和容错能力
- 通过添加更多实例实现横向扩展
- 根据需求使用不同的算法

## 配置

### 单服务模式（向后兼容）

现有的单服务配置无需修改即可继续工作：

```yaml
routes:
  - name: single-service
    path: /api
    backend:
      service:
        name: api.example.com
        port: 8080
        protocol: https
```

### 多服务模式

要启用负载均衡，使用 `servers` 字段配置多个服务器：

```yaml
routes:
  - name: load-balanced-service
    path: /api
    backend:
      service:
        algorithm: round-robin
        servers:
          - name: server1.example.com
            port: 8080
          - name: server2.example.com
            port: 8080
          - name: server3.example.com
            port: 8080
```

## 负载均衡算法

### Round-Robin（轮询）

在所有健康的服务器之间均匀轮询分配请求。

```yaml
backend:
  service:
    algorithm: round-robin
    servers:
      - name: server1.example.com
        port: 8080
      - name: server2.example.com
        port: 8080
      - name: server3.example.com
        port: 8080
```

**使用场景：**
- 所有服务器容量相似时
- 无状态服务
- 需要均匀分配流量

### Weighted（加权轮询）

根据服务器权重分配请求。权重更高的服务器接收更多流量。

```yaml
backend:
  service:
    algorithm: weighted
    servers:
      - name: server1.example.com
        port: 8080
        weight: 1    # 接收 25% 的流量
      - name: server2.example.com
        port: 8080
        weight: 2    # 接收 50% 的流量
      - name: server3.example.com
        port: 8080
        weight: 1    # 接收 25% 的流量
```

**使用场景：**
- 服务器容量不同时
- 渐进式流量迁移
- A/B 测试场景

### Least Connections（最少连接）

将请求路由到活动连接数最少的服务器。

```yaml
backend:
  service:
    algorithm: least-connections
    servers:
      - name: server1.example.com
        port: 8080
      - name: server2.example.com
        port: 8080
```

**使用场景：**
- 长连接场景
- 请求处理时间差异较大时
- WebSocket 连接

### IP Hash（IP 哈希）

基于客户端 IP 地址路由请求，确保同一客户端始终访问同一服务器。

```yaml
backend:
  service:
    algorithm: ip-hash
    servers:
      - name: server1.example.com
        port: 8080
      - name: server2.example.com
        port: 8080
```

**使用场景：**
- 会话持久化
- 需要粘性会话
- 缓存优化

**注意：** IP Hash 优先使用 `X-Forwarded-For` 或 `X-Real-IP` 头中的客户端 IP，否则回退到 `RemoteAddr`。

## 服务器配置

### 基本服务器字段

```yaml
servers:
  - name: server1.example.com    # 必需：服务器主机名或 IP
    port: 8080                    # 必需：服务器端口
    protocol: https              # 可选：协议 (http/https)，未设置时继承全局配置
    weight: 1                    # 可选：权重（用于 weighted 算法，默认：1）
    disabled: false              # 可选：禁用服务器（默认：false，服务器默认启用）
```

### 服务器级别配置覆盖

您可以在服务器级别覆盖全局配置：

```yaml
backend:
  service:
    algorithm: round-robin
    servers:
      - name: server1.example.com
        port: 8080
        # 使用全局配置
      
      - name: server2.example.com
        port: 8080
        # 覆盖请求头
        request:
          headers:
            X-Instance: server-2
        # 覆盖健康检查路径
        health_check:
          path: /custom-health
    
    # 全局配置（应用于所有服务器，除非被覆盖）
    request:
      headers:
        X-Service: my-service
    health_check:
      enable: true
      path: /health
```

## 健康检查

健康检查会自动从不健康的服务器中移除负载均衡池。

### 全局健康检查

```yaml
backend:
  service:
    algorithm: round-robin
    servers:
      - name: server1.example.com
        port: 8080
      - name: server2.example.com
        port: 8080
    health_check:
      enable: true
      method: GET
      path: /health
      status: [200]
      interval: 30    # 检查间隔（秒）
      timeout: 5      # 请求超时（秒）
```

### 服务器级别健康检查

```yaml
servers:
  - name: server1.example.com
    port: 8080
    health_check:
      enable: true
      path: /custom-health  # 覆盖全局健康检查路径
      interval: 60          # 覆盖全局间隔
```

### 健康检查选项

| 字段 | 类型 | 默认值 | 描述 |
|------|------|--------|------|
| `enable` | bool | false | 启用健康检查 |
| `method` | string | GET | 健康检查的 HTTP 方法 |
| `path` | string | /health | 健康检查端点路径 |
| `status` | array | [200] | 有效的 HTTP 状态码 |
| `interval` | int | 30 | 检查间隔（秒） |
| `timeout` | int | 5 | 请求超时（秒） |
| `ok` | bool | false | 始终认为健康（跳过检查） |

## 示例

### 完整示例：带健康检查的 Round-Robin

```yaml
routes:
  - name: api-service
    path: /api
    backend:
      service:
        algorithm: round-robin
        servers:
          - name: api1.example.com
            port: 8080
          - name: api2.example.com
            port: 8080
          - name: api3.example.com
            port: 8080
        request:
          headers:
            X-Service: api-service
            X-API-Version: v1
        response:
          headers:
            X-Powered-By: api-gateway
        health_check:
          enable: true
          method: GET
          path: /health
          status: [200, 201]
          interval: 30
          timeout: 5
```

### 示例：加权分配

```yaml
routes:
  - name: weighted-api
    path: /api/weighted
    backend:
      service:
        algorithm: weighted
        servers:
          - name: api-small.example.com
            port: 8080
            weight: 1    # 20% 的流量
          - name: api-medium.example.com
            port: 8080
            weight: 2    # 40% 的流量
          - name: api-large.example.com
            port: 8080
            weight: 2    # 40% 的流量
```

### 示例：使用 IP Hash 的会话持久化

```yaml
routes:
  - name: session-service
    path: /session
    backend:
      service:
        algorithm: ip-hash
        servers:
          - name: session1.example.com
            port: 8080
          - name: session2.example.com
            port: 8080
```

## 行为说明

### 服务器选择

1. 只考虑**健康**且**启用**的服务器
2. 如果没有健康的服务器，请求返回 503 错误
3. 健康检查异步运行，不会阻塞请求

### 配置合并

- 全局配置应用于所有服务器
- 服务器级别配置覆盖全局配置
- 合并的配置包括：
  - 请求头和查询参数
  - 响应头
  - 认证设置
  - 健康检查设置

### 向后兼容性

- 现有的单服务配置无需修改即可工作
- 单服务模式在内部自动转换为多服务模式
- 所有现有功能继续正常工作

## 最佳实践

1. **健康检查**：生产环境始终启用健康检查
2. **算法选择**：根据使用场景选择算法：
   - 使用 `round-robin` 实现均匀分配
   - 使用 `weighted` 处理不同容量的服务器
   - 使用 `least-connections` 处理长连接
   - 使用 `ip-hash` 实现会话持久化
3. **服务器权重**：根据服务器容量设置适当的权重
4. **监控**：监控服务器健康和流量分配
5. **渐进式发布**：使用加权分配实现渐进式流量迁移

## 故障排查

### 没有健康的服务器

如果所有服务器都不健康：
- 检查健康检查配置
- 验证健康检查端点是否可访问
- 查看健康检查日志

### 流量分配不均

- 对于加权算法，验证权重设置是否正确
- 检查是否有服务器被标记为不健康
- 验证所有服务器是否已启用

### IP Hash 的会话问题

- 确保 `X-Forwarded-For` 或 `X-Real-IP` 头设置正确
- 验证客户端 IP 是否正确提取
- 检查负载均衡器是否保留客户端 IP

## 下一步

- [配置说明](/zh/guide/configuration) - 完整的配置参考
- [健康检查](/zh/guide/health-check) - 健康检查配置
- [使用示例](/zh/guide/examples) - 更多配置示例
