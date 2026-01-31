# Quick Start

Get started with API Gateway in minutes.

## Installation

Install API Gateway using Go:

```bash
go install github.com/go-zoox/api-gateway/cmd/api-gateway@latest
```

Or download the latest release from [GitHub Releases](https://github.com/go-zoox/api-gateway/releases).

## Basic Configuration

Create a configuration file `config.yaml`:

```yaml
version: v1

port: 8080

routes:
  - name: example
    path: /api
    backend:
      service:
        protocol: https
        name: httpbin.org
        port: 443
```

## Run

Start the API Gateway:

```bash
api-gateway -c config.yaml
```

The gateway will start on port 8080. You can now access your backend service through the gateway:

```bash
curl http://localhost:8080/api/get
```

## What's Next?

- Learn more about [Configuration](/guide/configuration)
- Explore [Routing](/guide/routing) options
- Check out [Examples](/guide/examples) for more use cases
