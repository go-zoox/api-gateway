package loadbalancer

import (
	"net/http"
	"sync"

	"github.com/go-zoox/api-gateway/core/route"
	"github.com/go-zoox/api-gateway/core/service"
)

// LoadBalancer defines the interface for load balancing algorithms
type LoadBalancer interface {
	// Select chooses a server instance from the backend
	Select(req *http.Request, backend *route.NormalizedBackend) (*service.Server, error)

	// UpdateHealth updates the health status of a server (optional)
	UpdateHealth(server *service.Server, healthy bool)

	// OnRequestStart is called when a request starts (for least-connections)
	OnRequestStart(backend *route.NormalizedBackend, server *service.Server)

	// OnRequestEnd is called when a request ends (for least-connections)
	OnRequestEnd(backend *route.NormalizedBackend, server *service.Server)
}

// Manager manages load balancers and health checks
type Manager struct {
	loadbalancers map[string]LoadBalancer
	healthManager *HealthCheckManager
	mu            sync.RWMutex
}

// NewManager creates a new load balancer manager
func NewManager() *Manager {
	return &Manager{
		loadbalancers: make(map[string]LoadBalancer),
		healthManager: NewHealthCheckManager(),
	}
}

// GetLoadBalancer returns a load balancer for the specified algorithm
func (m *Manager) GetLoadBalancer(algorithm string) LoadBalancer {
	m.mu.RLock()
	lb, exists := m.loadbalancers[algorithm]
	m.mu.RUnlock()

	if exists {
		return lb
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Double check
	if lb, exists := m.loadbalancers[algorithm]; exists {
		return lb
	}

	// Create new load balancer based on algorithm
	switch algorithm {
	case "round-robin":
		lb = NewRoundRobin()
	case "weighted":
		lb = NewWeightedRoundRobin()
	case "least-connections":
		lb = NewLeastConnections()
	case "ip-hash":
		lb = NewIPHash()
	default:
		// Default to round-robin
		lb = NewRoundRobin()
	}

	m.loadbalancers[algorithm] = lb
	return lb
}

// StartHealthCheck starts health checking for a backend
func (m *Manager) StartHealthCheck(backend *route.NormalizedBackend) error {
	return m.healthManager.Start(backend)
}

// StopHealthCheck stops health checking for a backend
func (m *Manager) StopHealthCheck(backendID string) error {
	return m.healthManager.Stop(backendID)
}

// StopAllHealthChecks stops all health checks
func (m *Manager) StopAllHealthChecks() error {
	return m.healthManager.StopAll()
}

// GetHealthCheckManager returns the health check manager
func (m *Manager) GetHealthCheckManager() *HealthCheckManager {
	return m.healthManager
}
