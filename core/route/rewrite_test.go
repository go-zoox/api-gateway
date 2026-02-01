package route

import (
	"testing"

	"github.com/go-zoox/api-gateway/core/service"
)

func TestRewritePrefix(t *testing.T) {
	r := &Route{
		Name:     "Home",
		Path:     "/",
		PathType: "prefix",
	}

	paths := [][]string{
		{"/", "/"},
		{"/abc", "/abc"},
		{"/a/b/c", "/a/b/c"},
	}

	for _, p := range paths {
		request := p[0]
		expected := p[1]
		if expected != r.Rewrite(request) {
			t.Fatalf("expected %s, but got %s", expected, r.Rewrite(request))
		}
	}
}

func TestRewriteNoDuplicateRules(t *testing.T) {
	// Create a route with a backend
	r := &Route{
		Name:     "TestRoute",
		Path:     "/api",
		PathType: "prefix",
		Backend: Backend{
			Service: service.Service{
				Name:     "example.com",
				Port:     8080,
				Protocol: "http",
			},
		},
	}

	// Get the normalized backend to access baseConfig
	normalizedBackend := r.Backend.Normalize()
	if normalizedBackend == nil {
		t.Fatal("Normalize() should not return nil")
	}

	initialCount := len(normalizedBackend.BaseConfig.Request.Path.Rewrites)

	// Call Rewrite multiple times
	r.Rewrite("/api/test1")
	r.Rewrite("/api/test2")
	r.Rewrite("/api/test3")
	r.Rewrite("/api/test4")

	// Verify that only one rewrite rule was added (not 4)
	finalCount := len(normalizedBackend.BaseConfig.Request.Path.Rewrites)
	expectedCount := initialCount + 1

	if finalCount != expectedCount {
		t.Errorf("Expected %d rewrite rules after multiple calls, got %d. Rules: %v", expectedCount, finalCount, normalizedBackend.BaseConfig.Request.Path.Rewrites)
	}

	// Verify the rewrite rule is correct
	expectedRule := "^/api(.*)$:$1"
	found := false
	for _, rule := range normalizedBackend.BaseConfig.Request.Path.Rewrites {
		if rule == expectedRule {
			if found {
				t.Errorf("Duplicate rewrite rule found: %s", expectedRule)
			}
			found = true
		}
	}

	if !found {
		t.Errorf("Expected rewrite rule %s not found in rules: %v", expectedRule, normalizedBackend.BaseConfig.Request.Path.Rewrites)
	}
}
