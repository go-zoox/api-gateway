package route

import (
	"strings"

	"github.com/go-zoox/api-gateway/core/service"
)

type Service = service.Service

type Backend struct {
	Service service.Service `config:"service"`
}

type RateLimit struct {
	Enable    bool              `config:"enable"`
	Algorithm string            `config:"algorithm,default=token-bucket"` // token-bucket, leaky-bucket, fixed-window
	KeyType   string            `config:"key_type,default=ip"`            // ip, user, apikey, clientid, header
	KeyHeader string            `config:"key_header"`                     // when key_type=header, specify header name
	Limit     int64             `config:"limit"`                          // limit count
	Window    int64             `config:"window"`                         // time window in seconds
	Burst     int64             `config:"burst"`                          // burst capacity (only for token-bucket)
	Message   string            `config:"message,default=Too Many Requests"`
	Headers   map[string]string `config:"headers"` // custom response headers
}

// JSONAuditOutputFile configures the file sink when output.provider is file.
type JSONAuditOutputFile struct {
	// Path is the filesystem path; each audit record is one appended line (NDJSON).
	Path string `config:"path"`
}

// JSONAuditHTTPOutput configures the HTTP sink when output.provider is http.
type JSONAuditHTTPOutput struct {
	URL string `config:"url"`
	// Method defaults to POST if empty.
	Method string `config:"method,default=POST"`
	// Headers are optional extra request headers (e.g. Authorization).
	Headers map[string]string `config:"headers"`
	// TimeoutSeconds caps the outbound request (default 5; must be >0).
	TimeoutSeconds int64 `config:"timeout_seconds,default=5"`
}

// JSONAuditOutputDatabase configures the DB sink when output.provider is database.
type JSONAuditOutputDatabase struct {
	// Engine must be one of postgres, mysql, or sqlite.
	Engine string `config:"engine"`
	// DSN is the database connection string.
	DSN string `config:"dsn"`
	// Host is used to build DSN when set (higher priority than DSN).
	Host string `config:"host"`
	// Port is used to build DSN when Host is set.
	Port int64 `config:"port"`
	// Username is used to build DSN when Host is set.
	Username string `config:"username"`
	// Password is used to build DSN when Host is set.
	Password string `config:"password"`
	// DB is database name (postgres/mysql) or file path (sqlite) when Host/DB mode is used.
	DB string `config:"db"`
}

// JSONAuditOutput groups sink selection (provider) and provider-specific settings under json_audit.output.
type JSONAuditOutput struct {
	// Provider is console (default), file, http, or database.
	// Aliases: webhook/endpoint/api => http, db/sql => database.
	Provider string                  `config:"provider,default=console"`
	File     JSONAuditOutputFile     `config:"file"`
	HTTP     JSONAuditHTTPOutput     `config:"http"`
	Database JSONAuditOutputDatabase `config:"database"`
}

// JSONAudit configures JSON response audit logging for the gateway or a single route.
type JSONAudit struct {
	Enable bool `config:"enable"`
	// Output configures where each audit line is written (see provider, file, http).
	Output JSONAuditOutput `config:"output"`
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
	// Redact controls masking of sensitive headers, query keys, and JSON object keys in audit logs.
	Redact JSONAuditRedact `config:"redact"`
}

// JSONAuditRedact configures whether and how values are masked in audit records.
type JSONAuditRedact struct {
	// Enable turns redaction on or off. Omitted (nil) means on (default).
	Enable *bool `config:"enable"`
	// Keys lists JSON object keys and query parameter names to mask (case-insensitive).
	// Empty uses built-in defaults when redaction is enabled.
	Keys []string `config:"keys"`
}

// RedactEnabled reports whether masking is active. When Enable is omitted, defaults to true.
func (r JSONAuditRedact) RedactEnabled() bool {
	if r.Enable == nil {
		return true
	}
	return *r.Enable
}

// IPPolicy filters client IPs at the edge. Deny rules are evaluated first; then, if allow is non-empty,
// the client must fall into at least one allow CIDR; if allow is empty, only deny is applied.
type IPPolicy struct {
	Enable bool `config:"enable"`
	// Allow is a list of CIDRs. If non-empty, the client IP must match at least one entry.
	Allow []string `config:"allow"`
	// Deny is a list of CIDRs; matching clients receive HTTP 403.
	Deny []string `config:"deny"`
	// TrustedProxies lists CIDRs of reverse proxies. Only when the direct peer address is in this set
	// the gateway trusts X-Forwarded-For (first hop) to derive the client IP. Empty means the gateway
	// only uses the direct TCP remote address.
	TrustedProxies []string `config:"trusted_proxies"`
	// Message is the response body for denied requests.
	Message string `config:"message,default=Forbidden"`
}

// CORS adds Cross-Origin Resource Sharing headers. Enable at the global level and/or per route; route
// settings override the global block for fields that are set.
type CORS struct {
	Enable bool `config:"enable"`
	// AllowOrigins lists allowed Origin values; use * for any origin (incompatible with AllowCredentials).
	AllowOrigins []string `config:"allow_origins"`
	// AllowMethods lists allowed methods for preflight and Access-Control-Allow-Methods.
	AllowMethods []string `config:"allow_methods"`
	// AllowHeaders lists allowed request headers (Access-Control-Allow-Headers).
	AllowHeaders []string `config:"allow_headers"`
	// ExposeHeaders lists response headers the browser may read (Access-Control-Expose-Headers).
	ExposeHeaders    []string `config:"expose_headers"`
	AllowCredentials bool     `config:"allow_credentials"`
	// MaxAge is the preflight cache duration in seconds (Access-Control-Max-Age).
	MaxAge int64 `config:"max_age"`
}

type Route struct {
	Name    string  `config:"name"`
	Path    string  `config:"path"`
	Backend Backend `config:"backend"`
	// PathType is the path type of route, options: prefix, regex
	PathType  string    `config:"path_type,default=prefix"`
	RateLimit RateLimit `config:"rate_limit"`
	JSONAudit JSONAudit `config:"json_audit"`
	HTTPCache HTTPCache `config:"http_cache"`
	IPPolicy  IPPolicy  `config:"ip_policy"`
	CORS      CORS      `config:"cors"`
}

// EffectiveJSONAuditProvider returns the normalized sink id: console, file, or http.
func EffectiveJSONAuditProvider(o JSONAuditOutput) string {
	switch strings.ToLower(strings.TrimSpace(o.Provider)) {
	case "", "console", "stdout":
		return "console"
	case "file":
		return "file"
	case "http", "https", "webhook", "endpoint", "api":
		return "http"
	case "database", "db", "sql":
		return "database"
	default:
		return "console"
	}
}
