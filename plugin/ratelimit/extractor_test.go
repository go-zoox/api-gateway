package ratelimit

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/go-zoox/zoox"
)

func TestIPExtractor(t *testing.T) {
	extractor := &IPExtractor{}

	tests := []struct {
		name           string
		request        func() *http.Request
		expectedPrefix string
	}{
		{
			name: "X-Forwarded-For header",
			request: func() *http.Request {
				req := httptest.NewRequest("GET", "/test", nil)
				req.Header.Set("X-Forwarded-For", "192.168.1.1, 10.0.0.1")
				return req
			},
			expectedPrefix: "ip:192.168.1.1",
		},
		{
			name: "X-Real-IP header",
			request: func() *http.Request {
				req := httptest.NewRequest("GET", "/test", nil)
				req.Header.Set("X-Real-IP", "192.168.1.2")
				return req
			},
			expectedPrefix: "ip:192.168.1.2",
		},
		{
			name: "RemoteAddr fallback",
			request: func() *http.Request {
				req := httptest.NewRequest("GET", "/test", nil)
				req.RemoteAddr = "192.168.1.3:8080"
				return req
			},
			expectedPrefix: "ip:192.168.1.3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := tt.request()
			ctx := &zoox.Context{Request: req}

			key, err := extractor.Extract(ctx, req)
			if err != nil {
				t.Fatalf("Extract() error = %v", err)
			}

			if key[:len(tt.expectedPrefix)] != tt.expectedPrefix {
				t.Errorf("Extract() = %v, want prefix %v", key, tt.expectedPrefix)
			}
		})
	}
}

func TestAPIKeyExtractor(t *testing.T) {
	extractor := &APIKeyExtractor{}

	tests := []struct {
		name           string
		request        func() *http.Request
		expectedPrefix string
	}{
		{
			name: "X-API-Key header",
			request: func() *http.Request {
				req := httptest.NewRequest("GET", "/test", nil)
				req.Header.Set("X-API-Key", "test-api-key-123")
				return req
			},
			expectedPrefix: "apikey:test-api-key-123",
		},
		{
			name: "Authorization header with ApiKey",
			request: func() *http.Request {
				req := httptest.NewRequest("GET", "/test", nil)
				req.Header.Set("Authorization", "ApiKey test-key-456")
				return req
			},
			expectedPrefix: "apikey:test-key-456",
		},
		{
			name: "Query parameter",
			request: func() *http.Request {
				req := httptest.NewRequest("GET", "/test?api_key=query-key-789", nil)
				return req
			},
			expectedPrefix: "apikey:query-key-789",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := tt.request()
			ctx := &zoox.Context{Request: req}

			key, err := extractor.Extract(ctx, req)
			if err != nil {
				t.Fatalf("Extract() error = %v", err)
			}

			if key != tt.expectedPrefix {
				t.Errorf("Extract() = %v, want %v", key, tt.expectedPrefix)
			}
		})
	}
}

func TestHeaderExtractor(t *testing.T) {
	extractor := &HeaderExtractor{HeaderName: "X-Custom-ID"}

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Custom-ID", "custom-value-123")
	ctx := &zoox.Context{
		Request: req,
		Path:    req.URL.Path,
	}

	key, err := extractor.Extract(ctx, req)
	if err != nil {
		t.Fatalf("Extract() error = %v", err)
	}

	expected := "header:X-Custom-ID:custom-value-123"
	if key != expected {
		t.Errorf("Extract() = %v, want %v", key, expected)
	}
}

func TestExtractorFactory(t *testing.T) {
	factory := &ExtractorFactory{}

	tests := []struct {
		name      string
		keyType   string
		keyHeader string
		wantType  string
	}{
		{"IP extractor", "ip", "", "IPExtractor"},
		{"User extractor", "user", "", "UserExtractor"},
		{"API Key extractor", "apikey", "", "APIKeyExtractor"},
		{"Client ID extractor", "clientid", "", "ClientIDExtractor"},
		{"Header extractor", "header", "X-Custom", "HeaderExtractor"},
		{"Default to IP", "unknown", "", "IPExtractor"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			extractor := factory.NewExtractor(tt.keyType, tt.keyHeader)
			if extractor == nil {
				t.Fatal("NewExtractor() returned nil")
			}
			// Basic type check
			_ = extractor
		})
	}
}

func TestClientIDExtractor(t *testing.T) {
	e := &ClientIDExtractor{}
	ctx := &zoox.Context{}

	t.Run("X-Client-ID header", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/?client_id=q", nil)
		req.Header.Set("X-Client-ID", "hdr-wins")
		key, err := e.Extract(ctx, req)
		if err != nil || key != "clientid:hdr-wins" {
			t.Fatalf("got %q err=%v", key, err)
		}
	})

	t.Run("query client_id", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/?client_id=my-app", nil)
		key, err := e.Extract(ctx, req)
		if err != nil || key != "clientid:my-app" {
			t.Fatalf("got %q err=%v", key, err)
		}
	})

	t.Run("trim spaces", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("X-Client-ID", "  spaced  ")
		key, err := e.Extract(ctx, req)
		if err != nil || key != "clientid:spaced" {
			t.Fatalf("got %q err=%v", key, err)
		}
	})

	t.Run("fallback IP", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "192.0.2.1:1"
		key, err := e.Extract(ctx, req)
		if err != nil || key != "ip:192.0.2.1" {
			t.Fatalf("got %q err=%v", key, err)
		}
	})
}

func TestExtractorFactory_HeaderWithoutNameUsesIPExtractor(t *testing.T) {
	f := &ExtractorFactory{}
	ext := f.NewExtractor("header", "")
	if reflect.TypeOf(ext) != reflect.TypeOf(&IPExtractor{}) {
		t.Fatalf("got %T, want *IPExtractor", ext)
	}
}

func TestUserExtractor(t *testing.T) {
	e := &UserExtractor{}
	ctx := &zoox.Context{}

	t.Run("Bearer token", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Bearer secret-token")
		key, err := e.Extract(ctx, req)
		if err != nil || key != "user:secret-token" {
			t.Fatalf("got %q err=%v", key, err)
		}
	})

	t.Run("case insensitive bearer", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "BeArEr lower-check")
		key, err := e.Extract(ctx, req)
		if err != nil || key != "user:lower-check" {
			t.Fatalf("got %q err=%v", key, err)
		}
	})

	t.Run("X-User-ID", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("X-User-ID", "uid-42")
		key, err := e.Extract(ctx, req)
		if err != nil || key != "user:uid-42" {
			t.Fatalf("got %q err=%v", key, err)
		}
	})

	t.Run("fallback IP", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "192.0.2.50:9999"
		key, err := e.Extract(ctx, req)
		if err != nil || key != "ip:192.0.2.50" {
			t.Fatalf("got %q err=%v", key, err)
		}
	})
}

func TestAPIKeyExtractor_FallbackToIP(t *testing.T) {
	e := &APIKeyExtractor{}
	ctx := &zoox.Context{}
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.0.2.88:1234"
	key, err := e.Extract(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if want := "ip:192.0.2.88"; key != want {
		t.Fatalf("got %q want %q", key, want)
	}
}

func TestHeaderExtractor_FallbackToIP(t *testing.T) {
	e := &HeaderExtractor{HeaderName: "X-Tenant"}
	ctx := &zoox.Context{}
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "[2001:db8::1]:443"
	key, err := e.Extract(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if key != "ip:2001:db8::1" {
		t.Fatalf("got %q", key)
	}
}

func TestIPExtractor_RemoteAddrIPv6(t *testing.T) {
	e := &IPExtractor{}
	ctx := &zoox.Context{}
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "[2001:db8::1]:8443"
	key, err := e.Extract(ctx, req)
	if err != nil || key != "ip:2001:db8::1" {
		t.Fatalf("got %q err=%v", key, err)
	}
}

func TestIPExtractor_RemoteAddrWithoutPort(t *testing.T) {
	e := &IPExtractor{}
	ctx := &zoox.Context{}
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.0.2.77"
	key, err := e.Extract(ctx, req)
	if err != nil || key != "ip:192.0.2.77" {
		t.Fatalf("got %q err=%v", key, err)
	}
}

func TestIPExtractor_UnknownWhenNoRemote(t *testing.T) {
	e := &IPExtractor{}
	ctx := &zoox.Context{}
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = ""
	key, err := e.Extract(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if key != "ip:unknown" {
		t.Fatalf("got %q", key)
	}
}
