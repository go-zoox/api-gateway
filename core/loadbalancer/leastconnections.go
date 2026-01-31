package loadbalancer

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/go-zoox/api-gateway/core/route"
	"github.com/go-zoox/api-gateway/core/service"
)

// LeastConnections implements least-connections load balancing
type LeastConnections struct {
	connections map[string]map[string]int64 // backend ID -> server ID -> connection count
	mu          sync.RWMutex
}

// NewLeastConnections creates a new least-connections load balancer
func NewLeastConnections() *LeastConnections {
	return &LeastConnections{
		connections: make(map[string]map[string]int64),
	}
}

// Select chooses a server with the least number of active connections
func (lc *LeastConnections) Select(req *http.Request, backend *route.NormalizedBackend) (*service.Server, error) {
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

	lc.mu.Lock()
	defer lc.mu.Unlock()

	backendID := getBackendID(backend)

	// Initialize connection counts if needed
	if lc.connections[backendID] == nil {
		lc.connections[backendID] = make(map[string]int64)
	}

	// Find server with minimum connections
	var selected *service.Server
	minConnections := int64(-1)

	for _, server := range healthyServers {
		serverID := server.ID()
		connCount := lc.connections[backendID][serverID]

		if minConnections == -1 || connCount < minConnections {
			minConnections = connCount
			selected = server
		}
	}

	if selected == nil {
		return nil, fmt.Errorf("failed to select server")
	}

	// Note: Connection count is incremented in OnRequestStart(), not here
	// This ensures proper tracking when the request actually starts
	return selected, nil
}

// UpdateHealth is a no-op for least-connections
func (lc *LeastConnections) UpdateHealth(server *service.Server, healthy bool) {
	// Least-connections doesn't need to track health changes
}

// OnRequestStart increments the connection count
func (lc *LeastConnections) OnRequestStart(backend *route.NormalizedBackend, server *service.Server) {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	backendID := getBackendID(backend)
	if lc.connections[backendID] == nil {
		lc.connections[backendID] = make(map[string]int64)
	}
	lc.connections[backendID][server.ID()]++
}

// OnRequestEnd decrements the connection count
func (lc *LeastConnections) OnRequestEnd(backend *route.NormalizedBackend, server *service.Server) {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	backendID := getBackendID(backend)
	if lc.connections[backendID] != nil {
		if lc.connections[backendID][server.ID()] > 0 {
			lc.connections[backendID][server.ID()]--
		}
	}
}
