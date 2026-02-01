# api-gateway - A Easy, Powerful, Fexible API Gateway

[![PkgGoDev](https://pkg.go.dev/badge/github.com/go-zoox/api-gateway)](https://pkg.go.github.com/go-zoox/api-gateway-gateway)
[![Build Status](https://github.com/go-zoox/api-gateway/actions/workflows/release.yml/badge.svg?branch=master)](httpgithub.com/go-zoox/api-gateway-gateway/actions/workflows/release.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/go-zoox/api-gateway)](https://goreportcard.com/repgithub.com/go-zoox/api-gateway-gateway)
[![Coverage Status](https://coveralls.io/repos/github/go-zoox/api-gateway/badge.svg?branch=master)](https://coveralls.io/github/go-zoox/api-gateway?branch=master)
[![GitHub issues](https://img.shields.io/github/issues/go-zoox/api-gateway.svg)](https://github.com/go-zoox/api-gateway/issues)
[![Release](https://img.shields.io/github/tag/go-zoox/api-gateway.svg?label=Release)](https://github.com/go-zoox/api-gateway/tags)


## Installation
To install the package, run:
```bash
go install github.com/go-zoox/api-gateway/cmd/api-gateway@latest
```

## Quick Start

```bash
# start api-gateway, cached in memory, default udp port: 80
api-gateway

# start api-gateway with config (see conf/api-gateway.yml for more options)
api-gateway -c api-gateway.yml
```

## Configuration
See the [configuration file](conf/api-gateway.yml).

## Features

### Current Features
- ✅ **Simple Configuration**: YAML-based configuration
- ✅ **Flexible Routing**: Support for prefix and regex-based path matching
- ✅ **Path Rewriting**: Advanced path rewriting rules
- ✅ **Plugin System**: Extensible plugin architecture
- ✅ **Health Checks**: Built-in health check for gateway and backend services
- ✅ **Request/Response Transformation**: Header and query parameter modification
- ✅ **Cache Support**: Redis cache integration
- ✅ **Authentication Config**: Support for various authentication types (basic, bearer, jwt, oauth2, oidc)

### Planned Features (Roadmap)

See [TODO List](docs/TODO.md) for detailed development plan.

**High Priority (Core Features)**:
- 🔴 Load Balancing - Multi-instance backend support with various algorithms
- 🔴 Rate Limiting - API throttling and rate control
- 🔴 Timeout Control - Request timeout management
- 🔴 Retry Mechanism - Automatic retry with backoff strategies
- 🔴 Monitoring & Observability - Prometheus metrics, distributed tracing

**Medium Priority (Important Features)**:
- 🔴 Circuit Breaker - Fault tolerance and failure handling
- 🔴 CORS Support - Cross-origin resource sharing
- 🔴 Authentication Implementation - Complete auth/authz implementation
- 🔴 Request Validation - Input validation and schema checking
- 🔴 Response Caching - Response caching with TTL support

**Low Priority (Enhancement Features)**:
- 🔴 WebSocket Support - WebSocket proxy support
- 🔴 SSL/TLS Termination - TLS termination at gateway
- 🔴 Service Discovery - Kubernetes, Consul, etcd integration
- 🔴 Canary Deployment - Traffic splitting and A/B testing
- 🔴 Request/Response Body Transformation - JSON/XML transformation
- 🔴 API Versioning - Structured API version management
- 🔴 Access Logging - Structured request/response logging
- 🔴 Multi-Protocol Support - gRPC, GraphQL support

## Contributing

We welcome contributions! Please see our [TODO List](docs/TODO.md) for features we're planning to implement.

## License
GoZoox is released under the [MIT License](./LICENSE).
