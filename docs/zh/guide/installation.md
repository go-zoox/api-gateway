# 安装

API Gateway 可以通过多种方式安装。

## 前置要求

- Go 1.24.0 或更高版本（用于从源码构建）
- Docker（用于容器化部署）

## 使用 Go 安装

安装 API Gateway 最简单的方法是使用 Go：

```bash
go install github.com/go-zoox/api-gateway/cmd/api-gateway@latest
```

这会将 `api-gateway` 二进制文件安装到您的 `$GOPATH/bin` 目录。

## 从源码构建

克隆仓库：

```bash
git clone https://github.com/go-zoox/api-gateway.git
cd api-gateway
```

构建二进制文件：

```bash
go build -o api-gateway ./cmd/api-gateway
```

## Docker

拉取 Docker 镜像：

```bash
docker pull gozoox/api-gateway:latest
```

或从 Dockerfile 构建：

```bash
docker build -t api-gateway .
```

使用 Docker 运行：

```bash
docker run -d \
  -p 8080:8080 \
  -v $(pwd)/config.yaml:/etc/api-gateway/config.yaml \
  api-gateway
```

## Docker Compose

使用提供的 `docker-compose.yml`：

```bash
docker-compose up -d
```

## 验证安装

检查版本：

```bash
api-gateway --version
```

您应该看到版本号，例如：`1.4.5`

## 下一步

- [快速开始](/zh/guide/getting-started) - 创建您的第一个配置
- [配置](/zh/guide/configuration) - 了解所有配置选项
