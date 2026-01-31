# 快速开始

几分钟内开始使用 API Gateway。

## 安装

使用 Go 安装 API Gateway：

```bash
go install github.com/go-zoox/api-gateway/cmd/api-gateway@latest
```

或从 [GitHub Releases](https://github.com/go-zoox/api-gateway/releases) 下载最新版本。

## 基础配置

创建配置文件 `config.yaml`：

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

## 运行

启动 API Gateway：

```bash
api-gateway -c config.yaml
```

网关将在 8080 端口启动。现在您可以通过网关访问后端服务：

```bash
curl http://localhost:8080/api/get
```

## 下一步

- 了解更多关于[配置](/zh/guide/configuration)
- 探索[路由](/zh/guide/routing)选项
- 查看[示例](/zh/guide/examples)了解更多用例
