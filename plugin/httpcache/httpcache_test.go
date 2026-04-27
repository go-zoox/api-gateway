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

func TestBuildCacheKey_StableQueryOrder(t *testing.T) {
	h := &HTTPCachePlugin{}
	cfg := route.HTTPCache{}

	eq := func(a, b string) {
		t.Helper()
		ra, err := http.NewRequest(http.MethodGet, a, nil)
		if err != nil {
			t.Fatal(err)
		}
		rb, err := http.NewRequest(http.MethodGet, b, nil)
		if err != nil {
			t.Fatal(err)
		}
		ka, err := h.buildCacheKey(&cfg, ra, "/api")
		if err != nil {
			t.Fatal(err)
		}
		kb, err := h.buildCacheKey(&cfg, rb, "/api")
		if err != nil {
			t.Fatal(err)
		}
		if ka != kb {
			t.Fatalf("query order should not affect key: %q vs %q\n%q %q", ka, kb, a, b)
		}
	}
	eq("http://example.test/r?b=2&a=1", "http://example.test/r?a=1&b=2")
	eq("http://example.test/r?z=1&z=2&y=x", "http://example.test/r?y=x&z=2&z=1")
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

func TestBuildCacheKey_StableVaryHeaderOrder(t *testing.T) {
	h := &HTTPCachePlugin{}
	cfg := route.HTTPCache{
		VaryHeaders: []string{"Accept"},
	}
	req, err := http.NewRequest(http.MethodGet, "http://example.test/x", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header["Accept"] = []string{"text/html", "application/json", "text/plain"}
	k1, err := h.buildCacheKey(&cfg, req, "/")
	if err != nil {
		t.Fatal(err)
	}
	req2, err := http.NewRequest(http.MethodGet, "http://example.test/x", nil)
	if err != nil {
		t.Fatal(err)
	}
	req2.Header["Accept"] = []string{"application/json", "text/html", "text/plain"}
	k2, err := h.buildCacheKey(&cfg, req2, "/")
	if err != nil {
		t.Fatal(err)
	}
	if k1 != k2 {
		t.Fatalf("permutation of header values should not affect key: %q %q", k1, k2)
	}
}
