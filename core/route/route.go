package route

import (
	"github.com/go-zoox/api-gateway/core/service"
)

type Service = service.Service

type Backend struct {
	Service service.Service `config:"service"`
}

type Route struct {
	Name    string  `config:"name"`
	Path    string  `config:"path"`
	Backend Backend `config:"backend"`
	// PathType is the path type of route, options: prefix, regex
	PathType string `config:"path_type,default=prefix"`
}
