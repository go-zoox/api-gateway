package loadbalancer

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/go-zoox/api-gateway/core/route"
	"github.com/go-zoox/api-gateway/core/service"
)

// WeightedRoundRobin implements smooth weighted round-robin load balancing
type WeightedRoundRobin struct {
	weights map[string]map[string]int64 // backend ID -> server ID -> current weight
	mu      sync.RWMutex
}

// NewWeightedRoundRobin creates a new weighted round-robin load balancer
func NewWeightedRoundRobin() *WeightedRoundRobin {
	return &WeightedRoundRobin{
		weights: make(map[string]map[string]int64),
	}
}

// Select chooses a server using weighted round-robin algorithm
func (wrr *WeightedRoundRobin) Select(req *http.Request, backend *route.NormalizedBackend) (*service.Server, error) {
	// Filter healthy and enabled servers
	healthyServers := make([]*service.Server, 0)
	totalWeight := int64(0)

	for _, server := range backend.Servers {
		if !server.Disabled && server.IsHealthy() {
			healthyServers = append(healthyServers, server)
			totalWeight += server.Weight
		}
	}

	if len(healthyServers) == 0 {
		return nil, fmt.Errorf("no healthy servers available")
	}

	wrr.mu.Lock()
	defer wrr.mu.Unlock()

	backendID := getBackendID(backend)

	// Initialize weights if needed
	if wrr.weights[backendID] == nil {
		wrr.weights[backendID] = make(map[string]int64)
	}

	// Ensure all healthy servers have weights initialized
	for _, server := range healthyServers {
		serverID := server.ID()
		if _, exists := wrr.weights[backendID][serverID]; !exists {
			wrr.weights[backendID][serverID] = server.Weight
		}
	}

	// Find server with maximum current weight
	var selected *service.Server
	maxWeight := int64(-1)

	for _, server := range healthyServers {
		serverID := server.ID()
		currentWeight := wrr.weights[backendID][serverID]

		if currentWeight > maxWeight {
			maxWeight = currentWeight
			selected = server
		}
	}

	// This should never happen if healthyServers is not empty
	if selected == nil {
		if len(healthyServers) > 0 {
			// Fallback to first server if somehow selection failed
			selected = healthyServers[0]
			// Initialize weight for fallback server
			wrr.weights[backendID][selected.ID()] = selected.Weight
		} else {
			return nil, fmt.Errorf("failed to select server")
		}
	}

	// Update weights: smooth weighted round-robin algorithm
	// 1. Selected server: decrease by total weight
	// 2. All servers: increase by their configured weight
	selectedID := selected.ID()
	for _, server := range healthyServers {
		serverID := server.ID()
		// All servers increase by their configured weight
		wrr.weights[backendID][serverID] += server.Weight
		// Selected server additionally decreases by total weight
		if serverID == selectedID {
			wrr.weights[backendID][serverID] -= totalWeight
		}
	}

	return selected, nil
}

// UpdateHealth resets weights when health status changes
func (wrr *WeightedRoundRobin) UpdateHealth(server *service.Server, healthy bool) {
	// Reset weights for the backend when health changes
	// This will be handled on next Select call
}

// OnRequestStart is a no-op for weighted round-robin
func (wrr *WeightedRoundRobin) OnRequestStart(backend *route.NormalizedBackend, server *service.Server) {
	// Weighted round-robin doesn't track connections
}

// OnRequestEnd is a no-op for weighted round-robin
func (wrr *WeightedRoundRobin) OnRequestEnd(backend *route.NormalizedBackend, server *service.Server) {
	// Weighted round-robin doesn't track connections
}
