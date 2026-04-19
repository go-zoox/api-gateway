package jsonaudit

import (
	"net/http"
	"net/url"
	"testing"
)

func TestCloneHeadersRedacted(t *testing.T) {
	h := http.Header{}
	h.Set("Content-Type", "application/json")
	h.Set("Authorization", "Bearer secret")
	out := cloneHeadersRedacted(h)
	if out["Authorization"][0] != "[REDACTED]" {
		t.Fatalf("authorization not redacted: %#v", out["Authorization"])
	}
	if out["Content-Type"][0] != "application/json" {
		t.Fatal()
	}
}

func TestRedactQueryValues(t *testing.T) {
	q := url.Values{}
	q.Set("page", "1")
	q.Set("token", "abc")
	keys := map[string]struct{}{"token": {}}
	out := redactQueryValues(q, keys)
	if out["token"][0] != "[REDACTED]" || out["page"][0] != "1" {
		t.Fatalf("%#v", out)
	}
}
