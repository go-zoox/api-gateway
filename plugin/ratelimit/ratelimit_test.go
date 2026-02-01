package ratelimit

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-zoox/api-gateway/config"
	"github.com/go-zoox/api-gateway/core/route"
	"github.com/go-zoox/logger"
	"github.com/go-zoox/zoox"
	"github.com/go-zoox/zoox/defaults"
)

// createTestContext creates a test context with logger
func createTestContext(req *http.Request) *zoox.Context {
	ctx := &zoox.Context{
		Request: req,
		Path:    req.URL.Path,
		Logger:  logger.New(),
	}
	return ctx
}

func TestRateLimit_Disabled(t *testing.T) {
	plugin := New()
	app := defaults.Default()
	cfg := &config.Config{
		RateLimit: route.RateLimit{
			Enable: false,
		},
	}

	err := plugin.Prepare(app, cfg)
	if err != nil {
		t.Fatalf("Prepare() error = %v", err)
	}

	// Create a test request
	req := httptest.NewRequest("GET", "/test", nil)
	ctx := createTestContext(req)

	// Should not block when disabled
	err = plugin.OnRequest(ctx, req)
	if err != nil {
		t.Errorf("OnRequest() error = %v, want nil (disabled)", err)
	}
}

func TestRateLimit_GlobalConfig(t *testing.T) {
	plugin := New()
	app := defaults.Default()
	cfg := &config.Config{
		RateLimit: route.RateLimit{
			Enable:    true,
			Algorithm: "fixed-window",
			Storage:   "memory",
			KeyType:   "ip",
			Limit:     5,
			Window:    60,
		},
	}

	err := plugin.Prepare(app, cfg)
	if err != nil {
		t.Fatalf("Prepare() error = %v", err)
	}

	req := httptest.NewRequest("GET", "/test", nil)
	ctx := createTestContext(req)

	// First 5 requests should be allowed
	for i := 0; i < 5; i++ {
		err = plugin.OnRequest(ctx, req)
		if err != nil {
			t.Errorf("OnRequest() error = %v, want nil for request %d", err, i+1)
		}
	}

	// 6th request should be blocked
	err = plugin.OnRequest(ctx, req)
	if err == nil {
		t.Error("OnRequest() error = nil, want rate limit error")
	}
}

func TestRateLimit_RouteConfig(t *testing.T) {
	plugin := New()
	app := defaults.Default()
	cfg := &config.Config{
		Routes: []route.Route{
			{
				Path: "/api/users",
				RateLimit: route.RateLimit{
					Enable:    true,
					Algorithm: "fixed-window",
					Storage:   "memory",
					KeyType:   "ip",
					Limit:     3,
					Window:    60,
				},
			},
		},
	}

	err := plugin.Prepare(app, cfg)
	if err != nil {
		t.Fatalf("Prepare() error = %v", err)
	}

	req := httptest.NewRequest("GET", "/api/users", nil)
	ctx := createTestContext(req)

	// First 3 requests should be allowed
	for i := 0; i < 3; i++ {
		err = plugin.OnRequest(ctx, req)
		if err != nil {
			t.Errorf("OnRequest() error = %v, want nil for request %d", err, i+1)
		}
	}

	// 4th request should be blocked
	err = plugin.OnRequest(ctx, req)
	if err == nil {
		t.Error("OnRequest() error = nil, want rate limit error")
	}
}

func TestRateLimit_KeyExtraction(t *testing.T) {
	plugin := New()
	app := defaults.Default()
	cfg := &config.Config{
		RateLimit: route.RateLimit{
			Enable:    true,
			Algorithm: "fixed-window",
			Storage:   "memory",
			KeyType:   "apikey",
			Limit:     2,
			Window:    60,
		},
	}

	err := plugin.Prepare(app, cfg)
	if err != nil {
		t.Fatalf("Prepare() error = %v", err)
	}

	// Request with API key
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-API-Key", "test-key-123")
	ctx := createTestContext(req)

	// Should be allowed
	err = plugin.OnRequest(ctx, req)
	if err != nil {
		t.Errorf("OnRequest() error = %v, want nil", err)
	}

	// Different API key should have separate limit
	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.Header.Set("X-API-Key", "test-key-456")
	ctx2 := createTestContext(req2)

	err = plugin.OnRequest(ctx2, req2)
	if err != nil {
		t.Errorf("OnRequest() error = %v, want nil (different key)", err)
	}
}

func TestRateLimit_OnResponse(t *testing.T) {
	plugin := New()
	app := defaults.Default()
	cfg := &config.Config{
		RateLimit: route.RateLimit{
			Enable:    true,
			Algorithm: "fixed-window",
			Storage:   "memory",
			KeyType:   "ip",
			Limit:     5,
			Window:    60,
		},
	}

	err := plugin.Prepare(app, cfg)
	if err != nil {
		t.Fatalf("Prepare() error = %v", err)
	}

	req := httptest.NewRequest("GET", "/test", nil)
	ctx := createTestContext(req)

	// Make a request to populate context
	plugin.OnRequest(ctx, req)

	// Create response
	res := &http.Response{
		Header: make(http.Header),
	}

	// OnResponse should set headers
	err = plugin.OnResponse(ctx, res)
	if err != nil {
		t.Errorf("OnResponse() error = %v", err)
	}

	// Check if headers are set
	if res.Header.Get("X-RateLimit-Limit") == "" {
		t.Error("X-RateLimit-Limit header not set")
	}
	if res.Header.Get("X-RateLimit-Remaining") == "" {
		t.Error("X-RateLimit-Remaining header not set")
	}
	if res.Header.Get("X-RateLimit-Reset") == "" {
		t.Error("X-RateLimit-Reset header not set")
	}
}

func TestRateLimit_InvalidConfig(t *testing.T) {
	plugin := New()
	app := defaults.Default()
	cfg := &config.Config{
		RateLimit: route.RateLimit{
			Enable:    true,
			Algorithm: "fixed-window",
			Storage:   "memory",
			KeyType:   "ip",
			Limit:     0, // Invalid: limit is 0
			Window:    60,
		},
	}

	err := plugin.Prepare(app, cfg)
	if err != nil {
		t.Fatalf("Prepare() error = %v", err)
	}

	req := httptest.NewRequest("GET", "/test", nil)
	ctx := createTestContext(req)

	// Should skip rate limiting with invalid config
	err = plugin.OnRequest(ctx, req)
	if err != nil {
		t.Errorf("OnRequest() error = %v, want nil (invalid config should be skipped)", err)
	}
}
