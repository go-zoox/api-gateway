package loadbalancer

import (
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/go-zoox/api-gateway/core/route"
	"github.com/go-zoox/api-gateway/core/service"
)

// RoundRobin implements round-robin load balancing
type RoundRobin struct {
	indices map[string]*atomic.Uint64
	mu      sync.RWMutex
}

// NewRoundRobin creates a new round-robin load balancer
func NewRoundRobin() *RoundRobin {
	return &RoundRobin{
		indices: make(map[string]*atomic.Uint64),
	}
}

// Select chooses a server using round-robin algorithm
func (rr *RoundRobin) Select(req *http.Request, backend *route.NormalizedBackend) (*service.Server, error) {
		// Filter healthy and enabled servers
		healthyServers := make([]*service.Server, 0)
		for _, server := range backend.Servers {
			if !server.Disabled && server.IsHealthy() {
				healthyServers = append(healthyServers, server)
			}
		}

	if len(healthyServers) == 0 {
		return nil, fmt.Errorf("no healthy servers available")
	}

	// Get or create index for this backend
	backendID := getBackendID(backend)
	rr.mu.Lock()
	idx, exists := rr.indices[backendID]
	if !exists {
		idx = &atomic.Uint64{}
		rr.indices[backendID] = idx
	}
	rr.mu.Unlock()

	// Get current index and increment atomically
	current := idx.Add(1) - 1
	selected := healthyServers[int(current)%len(healthyServers)]

	return selected, nil
}

// UpdateHealth is a no-op for round-robin
func (rr *RoundRobin) UpdateHealth(server *service.Server, healthy bool) {
	// Round-robin doesn't need to track health changes
}

// OnRequestStart is a no-op for round-robin
func (rr *RoundRobin) OnRequestStart(backend *route.NormalizedBackend, server *service.Server) {
	// Round-robin doesn't track connections
}

// OnRequestEnd is a no-op for round-robin
func (rr *RoundRobin) OnRequestEnd(backend *route.NormalizedBackend, server *service.Server) {
	// Round-robin doesn't track connections
}

// getBackendID generates a unique ID for a backend
func getBackendID(backend *route.NormalizedBackend) string {
	if len(backend.Servers) == 0 {
		return "empty"
	}
	return backend.Servers[0].ID()
}
