# Health Check

API Gateway provides health check functionality for both the gateway itself and backend services.

## External Health Check

The external health check provides an endpoint that can be used to verify the gateway is running.

### Configuration

```yaml
healthcheck:
  outer:
    enable: true
    path: /healthz
    ok: true
```

### Options

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `enable` | bool | false | Enable the health check endpoint |
| `path` | string | /healthz | Path for the health check endpoint |
| `ok` | bool | true | Always return OK (skip actual checks) |

### Usage

When enabled, the gateway responds to health check requests:

```bash
curl http://localhost:8080/healthz
# Returns: ok
```

This is useful for:
- Load balancer health checks
- Kubernetes liveness/readiness probes
- Monitoring systems

### Example

```yaml
healthcheck:
  outer:
    enable: true
    path: /healthz
    ok: true
```

Kubernetes probe configuration:

```yaml
livenessProbe:
  httpGet:
    path: /healthz
    port: 8080
  initialDelaySeconds: 10
  periodSeconds: 5
```

## Internal Health Check

The internal health check monitors backend services to ensure they are available.

### Configuration

```yaml
healthcheck:
  inner:
    enable: true
    interval: 30
    timeout: 5
```

### Options

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `enable` | bool | false | Enable internal health checks |
| `interval` | int | 30 | Check interval in seconds |
| `timeout` | int | 5 | Request timeout in seconds |

### Service Health Check

You can configure health checks for individual services:

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

### Service Health Check Options

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `enable` | bool | false | Enable health check for this service |
| `method` | string | GET | HTTP method for health check |
| `path` | string | /health | Health check endpoint path |
| `status` | array | [200] | Valid HTTP status codes |
| `interval` | int | 30 | Check interval in seconds |
| `timeout` | int | 5 | Request timeout in seconds |
| `ok` | bool | false | Always consider service healthy (skip checks) |

### Health Check Behavior

When a service health check fails:
- The gateway may stop routing requests to that service
- Errors are logged for monitoring
- The gateway continues to operate for other services

## Health Check Examples

### Basic Configuration

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

### Service-Specific Health Check

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

### Kubernetes Integration

```yaml
# Gateway configuration
healthcheck:
  outer:
    enable: true
    path: /healthz

---
# Kubernetes deployment
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

## Monitoring

Health check status can be monitored through:
- Gateway logs
- External monitoring systems (Prometheus, etc.)
- Kubernetes events

## Best Practices

1. **Use Simple Endpoints**: Keep health check endpoints lightweight
2. **Appropriate Intervals**: Balance between responsiveness and overhead
3. **Timeout Configuration**: Set reasonable timeouts to avoid hanging requests
4. **Status Codes**: Use appropriate HTTP status codes for health status
5. **Separate Endpoints**: Use different endpoints for liveness and readiness if needed

## Next Steps

- [Configuration](/guide/configuration) - Complete configuration reference
- [Routing](/guide/routing) - Configure routes with health checks
- [Examples](/guide/examples) - See health check examples
