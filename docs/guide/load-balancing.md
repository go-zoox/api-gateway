# Load Balancing

API Gateway supports multiple load balancing algorithms to distribute traffic across backend service instances.

## Overview

Load balancing allows you to:
- Distribute traffic across multiple backend instances
- Improve availability and fault tolerance
- Scale horizontally by adding more instances
- Use different algorithms based on your needs

## Configuration

### Single Server Mode (Backward Compatible)

The existing single-server configuration continues to work without modification:

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

### Multi-Server Mode

To enable load balancing, configure multiple servers using the `servers` field:

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

## Load Balancing Algorithms

### Round-Robin

Distributes requests evenly across all healthy servers in rotation.

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

**Use Cases:**
- When all servers have similar capacity
- For stateless services
- When you want even distribution

### Weighted Round-Robin

Distributes requests based on server weights. Servers with higher weights receive more traffic.

```yaml
backend:
  service:
    algorithm: weighted
    servers:
      - name: server1.example.com
        port: 8080
        weight: 1    # Receives 25% of traffic
      - name: server2.example.com
        port: 8080
        weight: 2    # Receives 50% of traffic
      - name: server3.example.com
        port: 8080
        weight: 1    # Receives 25% of traffic
```

**Use Cases:**
- When servers have different capacities
- For gradual traffic migration
- For A/B testing scenarios

### Least Connections

Routes requests to the server with the fewest active connections.

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

**Use Cases:**
- For long-lived connections
- When request processing times vary
- For WebSocket connections

### IP Hash

Routes requests based on the client IP address, ensuring the same client always reaches the same server.

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

**Use Cases:**
- For session persistence
- When you need sticky sessions
- For caching optimization

**Note:** IP Hash uses the client IP from `X-Forwarded-For` or `X-Real-IP` headers if available, otherwise falls back to `RemoteAddr`.

## Server Configuration

### Basic Server Fields

```yaml
servers:
  - name: server1.example.com    # Required: Server hostname or IP
    port: 8080                    # Required: Server port
    protocol: https              # Optional: Protocol (http/https), inherits from base if not set
    weight: 1                    # Optional: Weight for weighted algorithm (default: 1)
    disabled: false              # Optional: Disable server (default: false, server is enabled by default)
```

### Server-Level Configuration Override

You can override global configuration at the server level:

```yaml
backend:
  service:
    algorithm: round-robin
    servers:
      - name: server1.example.com
        port: 8080
        # Uses global configuration
      
      - name: server2.example.com
        port: 8080
        # Override request headers
        request:
          headers:
            X-Instance: server-2
        # Override health check path
        health_check:
          path: /custom-health
    
    # Global configuration (applied to all servers unless overridden)
    request:
      headers:
        X-Service: my-service
    health_check:
      path: /health
```

## Health Checks

Health checks automatically remove unhealthy servers from the load balancing pool.

### Global Health Check

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
      method: GET
      path: /health
      status: [200]
      interval: 30    # Check interval in seconds
      timeout: 5      # Request timeout in seconds
```

### Server-Level Health Check

```yaml
servers:
  - name: server1.example.com
    port: 8080
    health_check:
      path: /custom-health  # Overrides global health check path
      interval: 60          # Overrides global interval
```

### Health Check Options

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `enable` | bool | false | Enable health checking |
| `method` | string | GET | HTTP method for health check |
| `path` | string | /health | Health check endpoint path |
| `status` | array | [200] | Valid HTTP status codes |
| `interval` | int | 30 | Check interval in seconds |
| `timeout` | int | 5 | Request timeout in seconds |
| `ok` | bool | false | Always consider healthy (skip checks) |

## Examples

### Complete Example: Round-Robin with Health Checks

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
          method: GET
          path: /health
          status: [200, 201]
          interval: 30
          timeout: 5
```

### Example: Weighted Distribution

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
            weight: 1    # 20% of traffic
          - name: api-medium.example.com
            port: 8080
            weight: 2    # 40% of traffic
          - name: api-large.example.com
            port: 8080
            weight: 2    # 40% of traffic
```

### Example: Session Persistence with IP Hash

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

## Behavior

### Server Selection

1. Only **healthy** and **enabled** servers are considered
2. If no healthy servers are available, the request returns a 503 error
3. Health checks run asynchronously and don't block requests

### Configuration Merging

- Global configuration is applied to all servers
- Server-level configuration overrides global configuration
- Merged configurations include:
  - Request headers and query parameters
  - Response headers
  - Authentication settings
  - Health check settings

### Backward Compatibility

- Existing single-server configurations work without modification
- Single-server mode is automatically converted to multi-server mode internally
- All existing features continue to work as before

## Best Practices

1. **Health Checks**: Always enable health checks for production deployments
2. **Algorithm Selection**: Choose the algorithm based on your use case:
   - Use `round-robin` for even distribution
   - Use `weighted` for servers with different capacities
   - Use `least-connections` for long-lived connections
   - Use `ip-hash` for session persistence
3. **Server Weights**: Set appropriate weights based on server capacity
4. **Monitoring**: Monitor server health and traffic distribution
5. **Gradual Rollout**: Use weighted distribution for gradual traffic migration

## Troubleshooting

### No Healthy Servers

If all servers are unhealthy:
- Check health check configuration
- Verify health check endpoints are accessible
- Review health check logs

### Uneven Traffic Distribution

- For weighted algorithm, verify weights are set correctly
- Check if some servers are being marked as unhealthy
- Verify all servers are not disabled

### Session Issues with IP Hash

- Ensure `X-Forwarded-For` or `X-Real-IP` headers are set correctly
- Verify client IPs are being extracted properly
- Check if load balancer is preserving client IPs

## Next Steps

- [Configuration](/guide/configuration) - Complete configuration reference
- [Health Check](/guide/health-check) - Health check configuration
- [Examples](/guide/examples) - More configuration examples
