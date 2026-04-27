package route

import (
	"net/http"
	"sort"
	"strings"
	"time"
)

// HTTPCache configures reverse-proxy style response caching for GET-style requests.
// Behavior is similar in spirit to nginx proxy_cache (custom keys, TTL, upstream Cache-Control).
type HTTPCache struct {
	Enable bool `config:"enable"`
	// TTL is entry time-to-live in seconds (default 60).
	TTL int64 `config:"ttl,default=60"`
	// KeyPrefix namespaces keys in Application.Cache() (default "httpcache").
	KeyPrefix string `config:"key_prefix"`
	// Methods lists HTTP methods to cache; empty means GET only.
	Methods []string `config:"methods"`
	// VaryHeaders lists request header names whose values are mixed into the cache key (RFC 7234 style).
	VaryHeaders []string `config:"vary_headers"`
	StatusMin int `config:"status_min,default=200"`
	StatusMax int `config:"status_max,default=299"`
	// MaxBodyBytes caps stored response bodies (default 2MiB).
	MaxBodyBytes int64 `config:"max_body_bytes,default=2097152"`
	// RespectUpstreamCacheControl skips storing when Cache-Control contains no-store or private.
	RespectUpstreamCacheControl bool `config:"respect_upstream_cache_control,default=true"`
	// CacheAuthorizedRequests allows caching when Authorization or Cookie is present (default false).
	CacheAuthorizedRequests bool `config:"cache_authorized_requests,default=false"`
	// SkipResponsesWithSetCookie avoids caching responses that set cookies (default true).
	SkipResponsesWithSetCookie bool `config:"skip_responses_with_set_cookie,default=true"`
	// OmitQueryFromKey excludes the raw query string from the cache key when true (default false).
	OmitQueryFromKey bool `config:"omit_query_from_key,default=false"`
}

// EffectiveTTL returns the KV TTL for stored entries.
func (h HTTPCache) EffectiveTTL() time.Duration {
	sec := h.TTL
	if sec <= 0 {
		sec = 60
	}
	return time.Duration(sec) * time.Second
}

// EffectiveMaxBodyBytes returns the max captured body size.
func (h HTTPCache) EffectiveMaxBodyBytes() int64 {
	if h.MaxBodyBytes <= 0 {
		return 2 * 1024 * 1024
	}
	return h.MaxBodyBytes
}

// EffectiveStatusRange returns inclusive HTTP status bounds for cacheable responses.
func (h HTTPCache) EffectiveStatusRange() (min, max int) {
	min, max = h.StatusMin, h.StatusMax
	if min <= 0 {
		min = 200
	}
	if max <= 0 {
		max = 299
	}
	return min, max
}

// IncludeQueryInKey reports whether the raw query string participates in the cache key.
func (h HTTPCache) IncludeQueryInKey() bool {
	return !h.OmitQueryFromKey
}

// NormalizedVaryHeaders returns sorted, canonical header names for stable cache keys.
func (h HTTPCache) NormalizedVaryHeaders() []string {
	if len(h.VaryHeaders) == 0 {
		return nil
	}
	out := make([]string, len(h.VaryHeaders))
	for i, v := range h.VaryHeaders {
		out[i] = http.CanonicalHeaderKey(strings.TrimSpace(v))
	}
	sort.Strings(out)
	return out
}
