package loadbalancer

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-zoox/api-gateway/core/route"
	"github.com/go-zoox/api-gateway/core/service"
	"github.com/go-zoox/logger"
)

// HealthCheckManager manages health checks for backend servers
type HealthCheckManager struct {
	checkers map[string]*healthChecker
	client   *http.Client
	mu       sync.RWMutex
}

// healthChecker represents a health check for a backend
type healthChecker struct {
	backend  *route.NormalizedBackend
	ctx      context.Context
	cancel   context.CancelFunc
	interval time.Duration
	timeout  time.Duration
}

// NewHealthCheckManager creates a new health check manager
func NewHealthCheckManager() *HealthCheckManager {
	return &HealthCheckManager{
		checkers: make(map[string]*healthChecker),
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// Start starts health checking for a backend
func (hcm *HealthCheckManager) Start(backend *route.NormalizedBackend) error {
	// Check if health check is enabled
	healthCheck := backend.BaseConfig.HealthCheck
	if !healthCheck.Enable || healthCheck.Ok {
		// Health check disabled or always OK
		return nil
	}

	// Set default interval and timeout if not specified
	interval := time.Duration(healthCheck.Interval) * time.Second
	if interval == 0 {
		interval = 30 * time.Second
	}

	timeout := time.Duration(healthCheck.Timeout) * time.Second
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	backendID := getBackendID(backend)

	hcm.mu.Lock()
	defer hcm.mu.Unlock()

	// Check if already started
	if _, exists := hcm.checkers[backendID]; exists {
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	checker := &healthChecker{
		backend:  backend,
		ctx:      ctx,
		cancel:   cancel,
		interval: interval,
		timeout:  timeout,
	}

	hcm.checkers[backendID] = checker

	// Start health check goroutine
	go checker.run(hcm.client)

	return nil
}

// Stop stops health checking for a backend
func (hcm *HealthCheckManager) Stop(backendID string) error {
	hcm.mu.Lock()
	defer hcm.mu.Unlock()

	checker, exists := hcm.checkers[backendID]
	if !exists {
		return nil
	}

	checker.cancel()
	delete(hcm.checkers, backendID)

	return nil
}

// StopAll stops all health checks
func (hcm *HealthCheckManager) StopAll() error {
	hcm.mu.Lock()
	defer hcm.mu.Unlock()

	for _, checker := range hcm.checkers {
		checker.cancel()
	}

	hcm.checkers = make(map[string]*healthChecker)

	return nil
}

// run executes the health check loop
func (hc *healthChecker) run(client *http.Client) {
	ticker := time.NewTicker(hc.interval)
	defer ticker.Stop()

	// Perform initial check immediately
	hc.checkAll(client)

	for {
		select {
		case <-hc.ctx.Done():
			return
		case <-ticker.C:
			hc.checkAll(client)
		}
	}
}

// checkAll checks all servers in the backend
func (hc *healthChecker) checkAll(client *http.Client) {
	for _, server := range hc.backend.Servers {
		if server.Disabled {
			server.SetHealthy(false) // Disabled servers are considered unhealthy
			continue
		}

		go hc.checkServer(server, client)
	}
}

// checkServer checks the health of a single server
func (hc *healthChecker) checkServer(server *service.Server, client *http.Client) {
	// Get health check configuration (server-level or backend-level)
	// Merge server override with base config to preserve base values when override has zero values
	healthCheck := hc.backend.BaseConfig.HealthCheck
	if server.HealthCheck != nil {
		override := *server.HealthCheck
		// Merge: only override fields that are explicitly set (non-zero/non-empty)
		// For Enable: only override if explicitly set to true (preserve base if false/zero-value)
		if override.Enable {
			healthCheck.Enable = override.Enable
		}
		if override.Method != "" {
			healthCheck.Method = override.Method
		}
		if override.Path != "" {
			healthCheck.Path = override.Path
		}
		if override.Status != nil {
			healthCheck.Status = override.Status
		}
		if override.Interval > 0 {
			healthCheck.Interval = override.Interval
		}
		if override.Timeout > 0 {
			healthCheck.Timeout = override.Timeout
		}
		if override.Ok {
			healthCheck.Ok = override.Ok
		}
	}

	// Skip if health check is disabled or always OK
	if !healthCheck.Enable || healthCheck.Ok {
		server.SetHealthy(true)
		return
	}

	// Build health check URL
	protocol := server.Protocol
	if protocol == "" {
		protocol = hc.backend.BaseConfig.Protocol
	}
	if protocol == "" {
		protocol = "http"
	}

	path := healthCheck.Path
	if path == "" {
		path = "/health"
	}

	method := healthCheck.Method
	if method == "" {
		method = "GET"
	}

	url := fmt.Sprintf("%s://%s%s", protocol, server.Host(), path)

	// Create request with timeout
	ctx, cancel := context.WithTimeout(hc.ctx, hc.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		logger.Errorf("[healthcheck] failed to create request for %s: %v", server.ID(), err)
		server.SetHealthy(false)
		return
	}

	// Perform health check
	resp, err := client.Do(req)
	if err != nil {
		logger.Debugf("[healthcheck] server %s is unhealthy: %v", server.ID(), err)
		server.SetHealthy(false)
		return
	}
	defer resp.Body.Close()

	// Check if status code is in the allowed list
	healthy := false
	for _, status := range healthCheck.Status {
		if int64(resp.StatusCode) == status {
			healthy = true
			break
		}
	}

	if healthy {
		logger.Debugf("[healthcheck] server %s is healthy (status: %d)", server.ID(), resp.StatusCode)
	} else {
		logger.Debugf("[healthcheck] server %s is unhealthy (status: %d, expected: %v)", server.ID(), resp.StatusCode, healthCheck.Status)
	}

	server.SetHealthy(healthy)
}
