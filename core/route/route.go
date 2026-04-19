package route

import (
	"github.com/go-zoox/api-gateway/core/service"
)

type Service = service.Service

type Backend struct {
	Service service.Service `config:"service"`
}

type RateLimit struct {
	Enable    bool              `config:"enable"`
	Algorithm string            `config:"algorithm,default=token-bucket"` // token-bucket, leaky-bucket, fixed-window
	KeyType   string            `config:"key_type,default=ip"`             // ip, user, apikey, clientid, header
	KeyHeader string            `config:"key_header"`                     // when key_type=header, specify header name
	Limit     int64             `config:"limit"`                          // limit count
	Window    int64             `config:"window"`                         // time window in seconds
	Burst     int64             `config:"burst"`                          // burst capacity (only for token-bucket)
	Message   string            `config:"message,default=Too Many Requests"`
	Headers   map[string]string `config:"headers"` // custom response headers
}

// JSONAudit configures JSON response audit logging for the gateway or a single route.
type JSONAudit struct {
	Enable bool `config:"enable"`
	// MaxBodyBytes caps captured request/response bodies (default 1MiB).
	MaxBodyBytes int64 `config:"max_body_bytes,default=1048576"`
	// SampleRate is the fraction of requests to audit (0.0–1.0]. Values <=0 are treated as 1.0.
	SampleRate float64 `config:"sample_rate,default=1"`
	// SniffJSON treats bodies as JSON when json.Valid succeeds if Content-Type is not JSON.
	SniffJSON bool `config:"sniff_json,default=true"`
	// DecompressGzip attempts gzip decompression for logging when Content-Encoding is gzip.
	DecompressGzip bool `config:"decompress_gzip,default=true"`
	// IncludePaths limits auditing to paths with these prefixes (empty = all, before excludes).
	IncludePaths []string `config:"include_paths"`
	// ExcludePaths skips paths with any of these prefixes.
	ExcludePaths []string `config:"exclude_paths"`
	// RedactKeys lists JSON object keys (any depth) to mask; empty uses built-in sensitive keys.
	RedactKeys []string `config:"redact_keys"`
}

type Route struct {
	Name    string  `config:"name"`
	Path    string  `config:"path"`
	Backend Backend `config:"backend"`
	// PathType is the path type of route, options: prefix, regex
	PathType   string    `config:"path_type,default=prefix"`
	RateLimit  RateLimit `config:"rate_limit"`
	JSONAudit  JSONAudit `config:"json_audit"`
}
