# 健康检查

API Gateway 为网关本身和后端服务提供健康检查功能。

## 外部健康检查

外部健康检查提供一个端点，可用于验证网关是否正在运行。

### 配置

```yaml
healthcheck:
  outer:
    enable: true
    path: /healthz
    ok: true
```

### 选项

| 字段 | 类型 | 默认值 | 描述 |
|------|------|--------|------|
| `enable` | bool | false | 启用健康检查端点 |
| `path` | string | /healthz | 健康检查端点路径 |
| `ok` | bool | true | 始终返回 OK（跳过实际检查） |

### 使用

启用后，网关响应健康检查请求：

```bash
curl http://localhost:8080/healthz
# 返回: ok
```

这可用于：
- 负载均衡器健康检查
- Kubernetes 存活/就绪探针
- 监控系统

### 示例

```yaml
healthcheck:
  outer:
    enable: true
    path: /healthz
    ok: true
```

Kubernetes 探针配置：

```yaml
livenessProbe:
  httpGet:
    path: /healthz
    port: 8080
  initialDelaySeconds: 10
  periodSeconds: 5
```

## 内部健康检查

内部健康检查监控后端服务以确保它们可用。

### 配置

```yaml
healthcheck:
  inner:
    enable: true
    interval: 30
    timeout: 5
```

### 选项

| 字段 | 类型 | 默认值 | 描述 |
|------|------|--------|------|
| `enable` | bool | false | 启用内部服务健康检查 |
| `interval` | int | 30 | 检查间隔（秒） |
| `timeout` | int | 5 | 请求超时（秒） |

### 服务健康检查

您可以为各个服务配置健康检查：

```yaml
routes:
  - name: api
    path: /api
    backend:
      service:
        name: api.example.com
        port: 8080
        health_check:
          enable: true
          method: GET
          path: /health
          status: [200]
          ok: false
```

### 服务健康检查选项

| 字段 | 类型 | 默认值 | 描述 |
|------|------|--------|------|
| `enable` | bool | false | 为此服务启用健康检查 |
| `method` | string | GET | 健康检查的 HTTP 方法 |
| `path` | string | /health | 健康检查端点路径 |
| `status` | array | [200] | 有效的 HTTP 状态码 |
| `ok` | bool | false | 始终认为服务健康（跳过检查） |

### 健康检查行为

当服务健康检查失败时：
- 网关可能停止将请求路由到该服务
- 错误被记录用于监控
- 网关继续为其他服务运行

## 健康检查示例

### 基础配置

```yaml
healthcheck:
  outer:
    enable: true
    path: /healthz
  inner:
    enable: true
    interval: 30
    timeout: 5
```

### 服务特定健康检查

```yaml
routes:
  - name: user-service
    path: /users
    backend:
      service:
        name: user-service.example.com
        port: 8080
        health_check:
          enable: true
          path: /api/health
          status: [200, 201]
```

### Kubernetes 集成

```yaml
# 网关配置
healthcheck:
  outer:
    enable: true
    path: /healthz

---
# Kubernetes 部署
apiVersion: apps/v1
kind: Deployment
metadata:
  name: api-gateway
spec:
  template:
    spec:
      containers:
      - name: api-gateway
        image: gozoox/api-gateway:latest
        livenessProbe:
          httpGet:
            path: /healthz
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 5
        readinessProbe:
          httpGet:
            path: /healthz
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 3
```

## 监控

健康检查状态可以通过以下方式监控：
- 网关日志
- 外部监控系统（Prometheus 等）
- Kubernetes 事件

## 最佳实践

1. **使用简单端点**：保持健康检查端点轻量级
2. **适当的间隔**：在响应性和开销之间取得平衡
3. **超时配置**：设置合理的超时以避免挂起请求
4. **状态码**：使用适当的 HTTP 状态码表示健康状态
5. **分离端点**：如果需要，为存活和就绪使用不同的端点

## 下一步

- [配置](/zh/guide/configuration) - 完整的配置参考
- [路由](/zh/guide/routing) - 配置带健康检查的路由
- [示例](/zh/guide/examples) - 查看健康检查示例
