package route

import (
	"github.com/go-zoox/api-gateway/core/service"
)

// NormalizedBackend represents a normalized backend configuration
// that can handle both single-server and multi-server modes
type NormalizedBackend struct {
	Algorithm  string
	Servers    []*service.Server
	BaseConfig *service.Service
}
