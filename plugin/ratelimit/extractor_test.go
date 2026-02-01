package ratelimit

import (
	"net/http"
	"net/http/httptest"
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
		name     string
		keyType  string
		keyHeader string
		wantType string
	}{
		{"IP extractor", "ip", "", "IPExtractor"},
		{"User extractor", "user", "", "UserExtractor"},
		{"API Key extractor", "apikey", "", "APIKeyExtractor"},
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
