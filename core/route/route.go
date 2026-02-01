package route

import (
	"github.com/go-zoox/api-gateway/core/service"
)

type Service = service.Service

type Backend struct {
	Service service.Service `config:"service"`
}

type RateLimit struct {
	Enable    bool              `config:"enable"`
	Algorithm string            `config:"algorithm,default=token-bucket"` // token-bucket, leaky-bucket, fixed-window
	Storage   string            `config:"storage,default=memory"`         // memory, redis
	KeyType   string            `config:"key_type,default=ip"`            // ip, user, apikey, header
	KeyHeader string            `config:"key_header"`                     // when key_type=header, specify header name
	Limit     int64             `config:"limit"`                          // limit count
	Window    int64             `config:"window"`                         // time window in seconds
	Burst     int64             `config:"burst"`                          // burst capacity (only for token-bucket)
	Message   string            `config:"message,default=Too Many Requests"`
	Headers   map[string]string `config:"headers"` // custom response headers
}

type Route struct {
	Name    string  `config:"name"`
	Path    string  `config:"path"`
	Backend Backend `config:"backend"`
	// PathType is the path type of route, options: prefix, regex
	PathType  string    `config:"path_type,default=prefix"`
	RateLimit RateLimit `config:"rate_limit"`
}
