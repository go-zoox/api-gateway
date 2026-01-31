# Installation

API Gateway can be installed in several ways.

## Prerequisites

- Go 1.24.0 or later (for building from source)
- Docker (for containerized deployment)

## Install from Go

The easiest way to install API Gateway is using Go:

```bash
go install github.com/go-zoox/api-gateway/cmd/api-gateway@latest
```

This will install the `api-gateway` binary to your `$GOPATH/bin` directory.

## Build from Source

Clone the repository:

```bash
git clone https://github.com/go-zoox/api-gateway.git
cd api-gateway
```

Build the binary:

```bash
go build -o api-gateway ./cmd/api-gateway
```

## Docker

Pull the Docker image:

```bash
docker pull gozoox/api-gateway:latest
```

Or build from Dockerfile:

```bash
docker build -t api-gateway .
```

Run with Docker:

```bash
docker run -d \
  -p 8080:8080 \
  -v $(pwd)/config.yaml:/etc/api-gateway/config.yaml \
  api-gateway
```

## Docker Compose

Use the provided `docker-compose.yml`:

```bash
docker-compose up -d
```

## Verify Installation

Check the version:

```bash
api-gateway --version
```

You should see the version number, for example: `1.4.5`

## Next Steps

- [Quick Start](/guide/getting-started) - Create your first configuration
- [Configuration](/guide/configuration) - Learn about all configuration options
