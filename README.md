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
sudo api-gateway

# start api-gateway with config (see conf/api-gateway.yml for more options)
sudo api-gateway -c api-gateway.yml
```

## Configuration
See the [configuration file](conf/api-gateway.yml).

## License
GoZoox is released under the [MIT License](./LICENSE).
