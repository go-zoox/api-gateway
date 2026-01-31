# Route API

The `Route` struct defines a route configuration for API Gateway.

## Type Definition

```go
type Route struct {
    Name     string  `config:"name"`
    Path     string  `config:"path"`
    Backend  Backend `config:"backend"`
    PathType string  `config:"path_type,default=prefix"`
}
```

## Fields

### Name

```go
Name string
```

Route name used for logging and identification. Required.

### Path

```go
Path string
```

Path pattern to match. Can be a prefix or regex pattern depending on `PathType`. Required.

### Backend

```go
Backend Backend
```

Backend service configuration. See [Backend Configuration](#backend-configuration). Required.

### PathType

```go
PathType string
```

Path matching type. Options:
- `prefix` (default): Prefix matching
- `regex`: Regular expression matching

## Backend Configuration

```go
type Backend struct {
    Service service.Service `config:"service"`
}
```

### Service

```go
type Service struct {
    Name        string   `config:"name"`
    Port        int64    `config:"port"`
    Protocol    string   `config:"protocol,default=http"`
    Request     Request  `config:"request"`
    Response    Response `config:"response"`
    Auth        Auth     `config:"auth"`
    HealthCheck HealthCheck `config:"health_check"`
}
```

#### Service Fields

- `Name` (string): Service hostname or IP address. Required.
- `Port` (int64): Service port. Default: `80`
- `Protocol` (string): Protocol (`http` or `https`). Default: `http`
- `Request` (Request): Request transformation configuration
- `Response` (Response): Response transformation configuration
- `Auth` (Auth): Authentication configuration
- `HealthCheck` (HealthCheck): Service-specific health check

## Request Configuration

```go
type Request struct {
    Path    RequestPath        `config:"path"`
    Headers map[string]string  `config:"headers"`
    Query   map[string]string  `config:"query"`
}
```

### Request Path

```go
type RequestPath struct {
    DisablePrefixRewrite bool     `config:"disable_prefix_rewrite"`
    Rewrites             []string `config:"rewrites"`
}
```

- `DisablePrefixRewrite` (bool): Disable automatic prefix removal. Default: `false`
- `Rewrites` ([]string): Path rewrite rules in format `pattern:replacement`

### Request Headers

```go
Headers map[string]string
```

Additional headers to add to requests.

### Request Query

```go
Query map[string]string
```

Additional query parameters to add to requests.

## Response Configuration

```go
type Response struct {
    Headers map[string]string `config:"headers"`
}
```

### Response Headers

```go
Headers map[string]string
```

Additional headers to add to responses.

## Authentication Configuration

```go
type Auth struct {
    Type         string   `config:"type"`
    Username     string   `config:"username"`
    Password     string   `config:"password"`
    Token        string   `config:"token"`
    Secret       string   `config:"secret"`
    Provider     string   `config:"provider"`
    ClientID     string   `config:"client_id"`
    ClientSecret string   `config:"client_secret"`
    RedirectURL  string   `config:"redirect_url"`
    Scopes       []string `config:"scopes"`
}
```

### Auth Types

- `basic`: Basic authentication (requires `username` and `password`)
- `bearer`: Bearer token authentication (requires `token`)
- `jwt`: JWT authentication (requires `secret`)
- `oauth2`: OAuth2 authentication (requires `provider`, `client_id`, `client_secret`, etc.)
- `oidc`: OpenID Connect authentication

## Service Health Check

```go
type HealthCheck struct {
    Enable bool    `config:"enable"`
    Method string  `config:"method,default=GET"`
    Path   string  `config:"path,default=/health"`
    Status []int64 `config:"status,default=[200]"`
    Ok     bool    `config:"ok"`
}
```

- `Enable` (bool): Enable health check for this service
- `Method` (string): HTTP method. Default: `GET`
- `Path` (string): Health check endpoint. Default: `/health`
- `Status` ([]int64): Valid HTTP status codes. Default: `[200]`
- `Ok` (bool): Always consider service healthy. Default: `false`

## Example

```go
route := route.Route{
    Name: "api",
    Path: "/api",
    PathType: "prefix",
    Backend: route.Backend{
        Service: service.Service{
            Name: "api.example.com",
            Port: 443,
            Protocol: "https",
            Request: service.Request{
                Path: service.RequestPath{
                    DisablePrefixRewrite: false,
                    Rewrites: []string{
                        "^/api/(.*):/$1",
                    },
                },
                Headers: map[string]string{
                    "X-API-Key": "secret",
                },
            },
            Response: service.Response{
                Headers: map[string]string{
                    "X-Powered-By": "api-gateway",
                },
            },
            HealthCheck: service.HealthCheck{
                Enable: true,
                Path: "/health",
                Status: []int64{200},
            },
        },
    },
}
```

## See Also

- [Config API](/api/config) - Main configuration
- [Plugin API](/api/plugin) - Plugin interface
- [Routing Guide](/guide/routing) - Routing guide
