package route

import (
	"github.com/go-zoox/api-gateway/core/service"
)

type Backend struct {
	Service service.Service `config:"service"`
}

type Route struct {
	Name    string  `config:"name"`
	Path    string  `config:"path"`
	Backend Backend `config:"backend"`
}
