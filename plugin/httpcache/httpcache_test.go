package httpcache

import (
	"net/http"
	"testing"

	"github.com/go-zoox/api-gateway/core/route"
)

func TestBuildCacheKey_StablePerVary(t *testing.T) {
	h := &HTTPCachePlugin{}
	cfg := route.HTTPCache{
		VaryHeaders: []string{"Accept-Language", "X-Api-Version"},
	}

	req1 := mustReq(t, http.MethodGet, "http://example.test/api/foo/bar?q=1", "en", "v2")
	req2 := mustReq(t, http.MethodGet, "http://example.test/api/foo/bar?q=1", "en", "v2")
	k1, err := h.buildCacheKey(&cfg, req1, "/api")
	if err != nil {
		t.Fatal(err)
	}
	k2, err := h.buildCacheKey(&cfg, req2, "/api")
	if err != nil {
		t.Fatal(err)
	}
	if k1 != k2 {
		t.Fatalf("same inputs should yield same key: %q vs %q", k1, k2)
	}

	req3 := mustReq(t, http.MethodGet, "http://example.test/api/foo/bar?q=1", "fr", "v2")
	k3, err := h.buildCacheKey(&cfg, req3, "/api")
	if err != nil {
		t.Fatal(err)
	}
	if k1 == k3 {
		t.Fatal("different vary header value should change key")
	}
}

func TestGetConfigForPath_LongestPrefix(t *testing.T) {
	h := &HTTPCachePlugin{
		global: &route.HTTPCache{Enable: true, TTL: 30},
		routes: map[string]*route.HTTPCache{
			"/api/v1": {Enable: true, TTL: 10},
			"/api":    {Enable: true, TTL: 20},
		},
	}

	cfg := h.getConfigForPath("/api/v1/users")
	if cfg == nil || cfg.TTL != 10 {
		t.Fatalf("want route /api/v1 ttl 10, got %+v", cfg)
	}

	cfg2 := h.getConfigForPath("/api/other")
	if cfg2 == nil || cfg2.TTL != 20 {
		t.Fatalf("want route /api ttl 20, got %+v", cfg2)
	}
}

func mustReq(t *testing.T, method, rawURL, lang, ver string) *http.Request {
	t.Helper()
	req, err := http.NewRequest(method, rawURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	if lang != "" {
		req.Header.Set("Accept-Language", lang)
	}
	if ver != "" {
		req.Header.Set("X-Api-Version", ver)
	}
	return req
}
