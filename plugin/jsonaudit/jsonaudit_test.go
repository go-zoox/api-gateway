package jsonaudit

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestJSONMIME(t *testing.T) {
	tests := []struct {
		ct   string
		want bool
	}{
		{"application/json", true},
		{"APPLICATION/JSON; charset=utf-8", true},
		{"application/vnd.api+json", true},
		{"text/plain", false},
		{"", false},
	}
	for _, tc := range tests {
		if got := jsonMIME(tc.ct); got != tc.want {
			t.Errorf("jsonMIME(%q) = %v, want %v", tc.ct, got, tc.want)
		}
	}
}

func TestHasPrefixAny(t *testing.T) {
	if !hasPrefixAny("/api/foo", []string{"/api"}) {
		t.Fatal("expected match")
	}
	if hasPrefixAny("/other", []string{"/api"}) {
		t.Fatal("unexpected match")
	}
	if !hasPrefixAny("/v1/a", []string{"/api", "/v1"}) {
		t.Fatal("expected second prefix")
	}
}

func TestRedactValue(t *testing.T) {
	keys := map[string]struct{}{
		"password": {}, "token": {},
	}
	raw := `{"user":"x","password":"secret","nested":{"token":"y"}}`
	var v any
	if err := json.Unmarshal([]byte(raw), &v); err != nil {
		t.Fatal(err)
	}
	out := redactValue(v, keys)
	b, err := json.Marshal(out)
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	if !strings.Contains(s, "[REDACTED]") || strings.Contains(s, "secret") {
		t.Fatalf("unexpected redacted output: %s", s)
	}
}

func TestReadBodyLimited(t *testing.T) {
	r := io.NopCloser(bytes.NewReader([]byte("hello")))
	b, trunc, err := readBodyLimited(r, 3)
	if err != nil {
		t.Fatal(err)
	}
	if trunc != true || string(b) != "hel" {
		t.Fatalf("got %q trunc=%v", b, trunc)
	}
}

func TestResponseLooksJSON(t *testing.T) {
	j := &JSONAudit{}
	j.cfg.SniffJSON = true

	if !j.responseLooksJSON("application/json", []byte(`not valid json`)) {
		t.Fatal("content-type json should qualify")
	}
	if j.responseLooksJSON("", []byte(`not json`)) {
		t.Fatal("non-json body should not qualify without sniff match")
	}
	if !j.responseLooksJSON("", []byte(`{"x":1}`)) {
		t.Fatal("sniff valid json")
	}
	j.cfg.SniffJSON = false
	if j.responseLooksJSON("", []byte(`{"x":1}`)) {
		t.Fatal("sniff disabled")
	}
}

func TestFirstHeader(t *testing.T) {
	req := &http.Request{Header: http.Header{}}
	req.Header.Set("X-Request-ID", "abc")
	if firstHeader(req, "X-Request-ID", "Other") != "abc" {
		t.Fatal()
	}
}
