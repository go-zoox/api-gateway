package route

import (
	"github.com/go-zoox/api-gateway/core/service"
)

// Normalize converts the Backend to a NormalizedBackend
// If only a single Service is configured, it automatically converts to multi-server mode
func (b *Backend) Normalize() *NormalizedBackend {
	// If servers are configured, use multi-server mode
	if len(b.Service.Servers) > 0 {
		servers := make([]*service.Server, 0, len(b.Service.Servers))
		for i := range b.Service.Servers {
			server := &b.Service.Servers[i]
			// Initialize health status to true by default
			server.SetHealthy(true)
			servers = append(servers, server)
		}

		algorithm := b.Service.Algorithm
		if algorithm == "" {
			algorithm = "round-robin"
		}

		return &NormalizedBackend{
			Algorithm:  algorithm,
			Servers:    servers,
			BaseConfig: &b.Service,
		}
	}

	// Backward compatibility: single server mode
	if b.Service.Name != "" {
		server := &service.Server{
			Name:     b.Service.Name,
			Port:     b.Service.Port,
			Protocol: b.Service.Protocol,
			Weight:   1,
			Disabled: false,
		}
		// Initialize health status to true by default
		server.SetHealthy(true)

		return &NormalizedBackend{
			Algorithm: "round-robin",
			Servers: []*service.Server{
				server,
			},
			BaseConfig: &b.Service,
		}
	}

	return nil
}

// IsMultiServer checks if the backend is configured for multi-server mode
func (b *Backend) IsMultiServer() bool {
	return len(b.Service.Servers) > 0
}

// GetService returns a single service (backward compatibility method)
// Returns the first server if in multi-server mode
func (b *Backend) GetService() *service.Service {
	if b.IsMultiServer() && len(b.Service.Servers) > 0 {
		// Return effective config of first server
		server := &b.Service.Servers[0]
		effective := server.GetEffectiveConfig(&b.Service)
		// Set name and port from server
		effective.Name = server.Name
		effective.Port = server.Port
		return effective
	}

	// Backward compatibility: return single service
	if b.Service.Name != "" {
		return &b.Service
	}

	return nil
}
