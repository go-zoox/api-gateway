package ratelimit

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-zoox/api-gateway/config"
	"github.com/go-zoox/api-gateway/core/route"
	"github.com/go-zoox/logger"
	"github.com/go-zoox/proxy"
	"github.com/go-zoox/zoox"
	"github.com/go-zoox/zoox/defaults"
)

// testZooxResponseWriter wraps httptest.ResponseRecorder so it satisfies zoox.ResponseWriter
// for tests that need Header() on the gateway response.
type testZooxResponseWriter struct {
	*httptest.ResponseRecorder
}

func (w *testZooxResponseWriter) CloseNotify() <-chan bool {
	ch := make(chan bool, 1)
	return ch
}

func (w *testZooxResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return nil, nil, fmt.Errorf("hijack not supported")
}

func (w *testZooxResponseWriter) Flush() {}

func (w *testZooxResponseWriter) Status() int {
	if w.Code != 0 {
		return w.Code
	}
	return http.StatusOK
}

func (w *testZooxResponseWriter) Size() int {
	return w.Body.Len()
}

func (w *testZooxResponseWriter) WriteString(s string) (int, error) {
	w.Write([]byte(s))
	return len(s), nil
}

func (w *testZooxResponseWriter) Written() bool {
	return w.Body.Len() > 0
}

func (w *testZooxResponseWriter) Pusher() http.Pusher {
	return nil
}

func (w *testZooxResponseWriter) WriteHeaderNow() {
	if w.Code == 0 {
		w.WriteHeader(http.StatusOK)
	}
}

var _ zoox.ResponseWriter = (*testZooxResponseWriter)(nil)

// createTestContext creates a test context with logger
func createTestContext(req *http.Request) *zoox.Context {
	ctx := &zoox.Context{
		Request: req,
		Path:    req.URL.Path,
		Logger:  logger.New(),
	}
	return ctx
}

func createTestContextWithWriter(req *http.Request, w zoox.ResponseWriter) *zoox.Context {
	ctx := &zoox.Context{
		Request:  req,
		Path:     req.URL.Path,
		Logger:   logger.New(),
		Writer:   w,
		Response: w,
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

func TestRateLimit_PathBoundaryMatching(t *testing.T) {
	// Test cases: path -> should match /users route?
	testCases := []struct {
		path        string
		shouldMatch bool
		description string
	}{
		{"/users", true, "exact match"},
		{"/users/", true, "exact match with trailing slash"},
		{"/users/123", true, "prefix match with /"},
		{"/users/test", true, "prefix match with /"},
		{"/users-admin/test", false, "should NOT match - prefix but next char is -"},
		{"/usersadmin", false, "should NOT match - prefix but no / boundary"},
		{"/user", false, "should NOT match - shorter path"},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			// Create a fresh plugin for each test case to avoid state sharing
			plugin := New()
			app := defaults.Default()
			cfg := &config.Config{
				Routes: []route.Route{
					{
						Path: "/users",
						RateLimit: route.RateLimit{
							Enable:    true,
							Algorithm: "fixed-window",
							Storage:   "memory",
							KeyType:   "ip",
							Limit:     5,
							Window:    60,
						},
					},
				},
			}

			err := plugin.Prepare(app, cfg)
			if err != nil {
				t.Fatalf("Prepare() error = %v", err)
			}

			// Make requests to see if rate limiting is applied
			// If route matches, 6th request should be blocked
			// If route doesn't match, all requests should pass (no rate limiting)
			for i := 0; i < 6; i++ {
				req := httptest.NewRequest("GET", tc.path, nil)
				ctx := createTestContext(req)
				err = plugin.OnRequest(ctx, req)

				if tc.shouldMatch {
					// Route should match, so 6th request should be blocked
					if i == 5 {
						if err == nil {
							t.Errorf("OnRequest() for path %s: expected rate limit error on 6th request, got nil", tc.path)
						}
					} else {
						if err != nil {
							t.Errorf("OnRequest() for path %s: unexpected error on request %d: %v", tc.path, i+1, err)
						}
					}
				} else {
					// Route should NOT match, so all requests should pass (no rate limiting applied)
					if err != nil {
						t.Errorf("OnRequest() for path %s: unexpected error (route should not match): %v", tc.path, err)
					}
				}
			}
		})
	}
}

func TestRateLimit_InvalidWindow(t *testing.T) {
	plugin := New()
	app := defaults.Default()
	cfg := &config.Config{
		RateLimit: route.RateLimit{
			Enable:    true,
			Algorithm: "fixed-window",
			Storage:   "memory",
			KeyType:   "ip",
			Limit:     10,
			Window:    0,
		},
	}

	if err := plugin.Prepare(app, cfg); err != nil {
		t.Fatalf("Prepare() error = %v", err)
	}

	req := httptest.NewRequest("GET", "/test", nil)
	ctx := createTestContext(req)
	if err := plugin.OnRequest(ctx, req); err != nil {
		t.Errorf("OnRequest() error = %v, want nil when window is invalid", err)
	}
}

func TestRateLimit_RouteOverridesGlobal(t *testing.T) {
	plugin := New()
	app := defaults.Default()
	cfg := &config.Config{
		RateLimit: route.RateLimit{
			Enable:    true,
			Algorithm: "fixed-window",
			Storage:   "memory",
			KeyType:   "ip",
			Limit:     100,
			Window:    60,
		},
		Routes: []route.Route{
			{
				Path: "/vip",
				RateLimit: route.RateLimit{
					Enable:    true,
					Algorithm: "fixed-window",
					Storage:   "memory",
					KeyType:   "ip",
					Limit:     2,
					Window:    60,
				},
			},
		},
	}

	if err := plugin.Prepare(app, cfg); err != nil {
		t.Fatalf("Prepare() error = %v", err)
	}

	vip := httptest.NewRequest("GET", "/vip", nil)
	vipCtx := createTestContext(vip)
	for i := 0; i < 2; i++ {
		if err := plugin.OnRequest(vipCtx, vip); err != nil {
			t.Fatalf("OnRequest() /vip request %d: %v", i+1, err)
		}
	}
	if err := plugin.OnRequest(vipCtx, vip); err == nil {
		t.Error("expected route-specific limit (2) to block the 3rd request to /vip")
	}

	other := httptest.NewRequest("GET", "/global-only", nil)
	otherCtx := createTestContext(other)
	for i := 0; i < 3; i++ {
		if err := plugin.OnRequest(otherCtx, other); err != nil {
			t.Fatalf("OnRequest() /global-only request %d: %v", i+1, err)
		}
	}
}

func TestRateLimit_LongestRoutePrefixWins(t *testing.T) {
	plugin := New()
	app := defaults.Default()
	cfg := &config.Config{
		Routes: []route.Route{
			{
				Path: "/api",
				RateLimit: route.RateLimit{
					Enable:    true,
					Algorithm: "fixed-window",
					Storage:   "memory",
					KeyType:   "ip",
					Limit:     50,
					Window:    60,
				},
			},
			{
				Path: "/api/v1",
				RateLimit: route.RateLimit{
					Enable:    true,
					Algorithm: "fixed-window",
					Storage:   "memory",
					KeyType:   "ip",
					Limit:     2,
					Window:    60,
				},
			},
		},
	}

	if err := plugin.Prepare(app, cfg); err != nil {
		t.Fatalf("Prepare() error = %v", err)
	}

	// /api/v1/foo must use /api/v1 (limit 2), not the looser /api rule.
	v1 := httptest.NewRequest("GET", "/api/v1/foo", nil)
	v1Ctx := createTestContext(v1)
	for i := 0; i < 2; i++ {
		if err := plugin.OnRequest(v1Ctx, v1); err != nil {
			t.Fatalf("OnRequest() /api/v1/foo request %d: %v", i+1, err)
		}
	}
	if err := plugin.OnRequest(v1Ctx, v1); err == nil {
		t.Error("expected /api/v1 rule to block the 3rd request to /api/v1/foo")
	}

	// /api/other matches /api only (not /api/v1).
	apiOther := httptest.NewRequest("GET", "/api/other", nil)
	apiOtherCtx := createTestContext(apiOther)
	for i := 0; i < 3; i++ {
		if err := plugin.OnRequest(apiOtherCtx, apiOther); err != nil {
			t.Fatalf("OnRequest() /api/other request %d: %v", i+1, err)
		}
	}
}

func TestRateLimit_ExceededCustomMessageAndHeaders(t *testing.T) {
	plugin := New()
	app := defaults.Default()
	cfg := &config.Config{
		RateLimit: route.RateLimit{
			Enable:    true,
			Algorithm: "fixed-window",
			Storage:   "memory",
			KeyType:   "ip",
			Limit:     1,
			Window:    60,
			Message:   "custom too many",
			Headers: map[string]string{
				"X-Rate-Policy": "test",
			},
		},
	}

	if err := plugin.Prepare(app, cfg); err != nil {
		t.Fatalf("Prepare() error = %v", err)
	}

	rec := httptest.NewRecorder()
	zw := &testZooxResponseWriter{ResponseRecorder: rec}
	req := httptest.NewRequest("GET", "/test", nil)
	ctx := createTestContextWithWriter(req, zw)

	if err := plugin.OnRequest(ctx, req); err != nil {
		t.Fatalf("first OnRequest: %v", err)
	}
	err := plugin.OnRequest(ctx, req)
	if err == nil {
		t.Fatal("expected second request to be rate limited")
	}

	var he *proxy.HTTPError
	if !errors.As(err, &he) {
		t.Fatalf("expected *proxy.HTTPError, got %T", err)
	}
	if he.Status() != http.StatusTooManyRequests {
		t.Errorf("status = %d, want %d", he.Status(), http.StatusTooManyRequests)
	}
	if he.Error() != "custom too many" {
		t.Errorf("message = %q, want %q", he.Error(), "custom too many")
	}

	if got := rec.Header().Get("X-RateLimit-Limit"); got != "1" {
		t.Errorf("X-RateLimit-Limit = %q, want %q", got, "1")
	}
	if rec.Header().Get("Retry-After") == "" {
		t.Error("expected Retry-After header on 429 response")
	}
	if rec.Header().Get("X-Rate-Policy") != "test" {
		t.Errorf("custom header X-Rate-Policy = %q, want %q", rec.Header().Get("X-Rate-Policy"), "test")
	}
}
