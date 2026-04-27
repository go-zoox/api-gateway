package ippolicy

import (
	"net/http"
	"net/netip"
	"testing"
)

func TestCompiled_allows(t *testing.T) {
	c, err := newCompiled(
		[]string{"10.0.0.0/8"},
		[]string{"10.0.0.1"},
		nil,
		"no",
	)
	if err != nil {
		t.Fatal(err)
	}
	a, _ := netip.ParseAddr("10.0.0.2")
	if !c.allows(a) {
		t.Fatalf("10.0.0.2 should be allowed in allowlist")
	}
	bad, _ := netip.ParseAddr("10.0.0.1")
	if c.allows(bad) {
		t.Fatalf("10.0.0.1 should be denied in deny list")
	}
}

func TestClientIP_usesXFFWhenTrusted(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.1:1234"
	p, _ := netip.ParsePrefix("10.0.0.0/8")
	direct, _ := DirectPeerIP(req)
	req.Header.Set("X-Forwarded-For", "198.51.100.1, 10.0.0.1")
	c := ClientIP(req, direct, []netip.Prefix{p})
	if c.String() != "198.51.100.1" {
		t.Fatalf("got %s", c)
	}
}

func TestClientIP_ignoresXFFWhenNotTrusted(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "198.51.100.1:1234"
	direct, _ := DirectPeerIP(req)
	p, _ := netip.ParsePrefix("10.0.0.0/8")
	req.Header.Set("X-Forwarded-For", "8.8.8.8")
	c := ClientIP(req, direct, []netip.Prefix{p})
	if c.String() != "198.51.100.1" {
		t.Fatalf("expected direct peer, got %s", c)
	}
}
