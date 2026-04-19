package config

import (
	"github.com/go-zoox/api-gateway/core/route"
)

type Config struct {
	Port int64 `config:"port"`
	// BaseURI is the base uri of api gateway, which is used to generate the proxy request url
	BaseURI string `config:"baseuri"`
	//
	Backend route.Backend `config:"backend"`
	//
	Routes []route.Route `config:"routes"`
	//
	Cache Cache `config:"cache"`
	//
	HealthCheck HealthCheck `config:"healthcheck"`
	//
	RateLimit RateLimit `config:"rate_limit"`
	//
	JSONAudit JSONAudit `config:"json_audit"`
	//
	// Match func(path string) (r *route.Route, err error)
}

// JSONAudit is the json_audit YAML section (also used by plugin/jsonaudit).
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

type HealthCheck struct {
	Outer HealthCheckOuter `config:"outer"`
	Inner HealthCheckInner `config:"inner"`
}

type HealthCheckOuter struct {
	Enable bool `config:"enable"`
	// Path is the health check request path
	Path string `config:"path"`
	// Ok means all health check request returns ok
	Ok bool `config:"ok"`
}

type HealthCheckInner struct {
	Enable bool `config:"enable"`
	//
	Interval int64 `config:"interval"`
	Timeout  int64 `config:"timeout"`
}

type Cache struct {
	Host     string `config:"host"`
	Port     int64  `config:"port"`
	Username string `config:"username"`
	Password string `config:"password"`
	DB       int64  `config:"db"`
	Prefix   string `config:"prefix"`
}

type SSL struct {
	Domain string  `config:"domain"`
	Cert   SSLCert `config:"cert"`
}

type SSLCert struct {
	Certificate    string `config:"certificate"`
	CertificateKey string `config:"certificate_key"`
}

// RateLimit uses route.RateLimit to avoid circular dependency
type RateLimit = route.RateLimit
