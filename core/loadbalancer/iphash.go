package loadbalancer

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/go-zoox/api-gateway/core/route"
	"github.com/go-zoox/api-gateway/core/service"
)

// IPHash implements IP hash load balancing
type IPHash struct{}

// NewIPHash creates a new IP hash load balancer
func NewIPHash() *IPHash {
	return &IPHash{}
}

// Select chooses a server based on client IP hash
func (ih *IPHash) Select(req *http.Request, backend *route.NormalizedBackend) (*service.Server, error) {
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

	// Get client IP
	clientIP := getClientIP(req)

	// Calculate hash from IP
	hash := sha256.Sum256([]byte(clientIP))
	hashValue := binary.BigEndian.Uint64(hash[:8])

	// Select server based on hash
	if len(healthyServers) == 0 {
		return nil, fmt.Errorf("no healthy servers available")
	}
	idx := int(hashValue) % len(healthyServers)
	if idx < 0 {
		idx = -idx
	}
	return healthyServers[idx], nil
}

// UpdateHealth is a no-op for IP hash
func (ih *IPHash) UpdateHealth(server *service.Server, healthy bool) {
	// IP hash doesn't need to track health changes
}

// OnRequestStart is a no-op for IP hash
func (ih *IPHash) OnRequestStart(backend *route.NormalizedBackend, server *service.Server) {
	// IP hash doesn't track connections
}

// OnRequestEnd is a no-op for IP hash
func (ih *IPHash) OnRequestEnd(backend *route.NormalizedBackend, server *service.Server) {
	// IP hash doesn't track connections
}

// getClientIP extracts the client IP from the request
func getClientIP(req *http.Request) string {
	// Try X-Forwarded-For first
	if xff := req.Header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Try X-Real-IP
	if xri := req.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	// Fall back to RemoteAddr
	ip, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		return req.RemoteAddr
	}
	return ip
}
