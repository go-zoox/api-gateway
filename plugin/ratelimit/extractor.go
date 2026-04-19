package ratelimit

import (
	"net"
	"net/http"
	"strings"

	"github.com/go-zoox/zoox"
)

// KeyExtractor extracts a key from the request for rate limiting
type KeyExtractor interface {
	Extract(ctx *zoox.Context, req *http.Request) (string, error)
}

// ExtractorFactory creates key extractors based on configuration
type ExtractorFactory struct{}

// NewExtractor creates a key extractor based on key type
func (f *ExtractorFactory) NewExtractor(keyType, keyHeader string) KeyExtractor {
	switch keyType {
	case "ip":
		return &IPExtractor{}
	case "user":
		return &UserExtractor{}
	case "apikey":
		return &APIKeyExtractor{}
	case "header":
		if keyHeader == "" {
			return &IPExtractor{} // fallback to IP if header not specified
		}
		return &HeaderExtractor{HeaderName: keyHeader}
	default:
		return &IPExtractor{} // default to IP
	}
}

// IPExtractor extracts the client IP address
type IPExtractor struct{}

func (e *IPExtractor) Extract(ctx *zoox.Context, req *http.Request) (string, error) {
	// Try X-Forwarded-For first (for proxies/load balancers)
	forwardedFor := req.Header.Get("X-Forwarded-For")
	if forwardedFor != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		ips := strings.Split(forwardedFor, ",")
		if len(ips) > 0 {
			ip := strings.TrimSpace(ips[0])
			if ip != "" {
				return "ip:" + ip, nil
			}
		}
	}

	// Try X-Real-IP
	realIP := req.Header.Get("X-Real-IP")
	if realIP != "" {
		return "ip:" + realIP, nil
	}

	// Fallback to RemoteAddr
	remoteAddr := req.RemoteAddr
	if remoteAddr != "" {
		// Remove port if present using net.SplitHostPort for proper IPv6 handling
		host, _, err := net.SplitHostPort(remoteAddr)
		if err == nil {
			remoteAddr = host
		}
		return "ip:" + remoteAddr, nil
	}

	return "ip:unknown", nil
}

// UserExtractor extracts user ID from JWT token or session
type UserExtractor struct{}

func (e *UserExtractor) Extract(ctx *zoox.Context, req *http.Request) (string, error) {
	// Try to extract from Authorization header (Bearer token)
	authHeader := req.Header.Get("Authorization")
	if authHeader != "" {
		// Check if it's a Bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
			// For now, use the token itself as the key
			// In a real implementation, you might want to decode JWT and extract user ID
			return "user:" + parts[1], nil
		}
	}

	// Try to extract from X-User-ID header
	userID := req.Header.Get("X-User-ID")
	if userID != "" {
		return "user:" + userID, nil
	}

	// Fallback to IP if no user identifier found
	ipExtractor := &IPExtractor{}
	return ipExtractor.Extract(ctx, req)
}

// APIKeyExtractor extracts API key from header or query parameter
type APIKeyExtractor struct{}

func (e *APIKeyExtractor) Extract(ctx *zoox.Context, req *http.Request) (string, error) {
	// Try X-API-Key header first
	apiKey := req.Header.Get("X-API-Key")
	if apiKey != "" {
		return "apikey:" + apiKey, nil
	}

	// Try Authorization header with ApiKey prefix
	authHeader := req.Header.Get("Authorization")
	if authHeader != "" {
		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && strings.ToLower(parts[0]) == "apikey" {
			return "apikey:" + parts[1], nil
		}
	}

	// Try query parameter
	apiKey = req.URL.Query().Get("api_key")
	if apiKey != "" {
		return "apikey:" + apiKey, nil
	}

	// Fallback to IP if no API key found
	ipExtractor := &IPExtractor{}
	return ipExtractor.Extract(ctx, req)
}

// HeaderExtractor extracts value from a custom header
type HeaderExtractor struct {
	HeaderName string
}

func (e *HeaderExtractor) Extract(ctx *zoox.Context, req *http.Request) (string, error) {
	value := req.Header.Get(e.HeaderName)
	if value != "" {
		return "header:" + e.HeaderName + ":" + value, nil
	}

	// Fallback to IP if header not found
	ipExtractor := &IPExtractor{}
	return ipExtractor.Extract(ctx, req)
}
