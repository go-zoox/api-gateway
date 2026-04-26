package cors

import "testing"

func TestAllowOriginValue_wildcard(t *testing.T) {
	c := &corsConfig{allowOrigins: []string{"*"}, allowCredentials: false}
	if allowOriginValue(c, "https://a.com") != "*" {
		t.Fatal()
	}
}

func TestAllowOriginValue_credentialReflection(t *testing.T) {
	c := &corsConfig{
		allowOrigins:     []string{"https://a.com", "https://b.com"},
		allowCredentials: true,
	}
	if allowOriginValue(c, "https://a.com") != "https://a.com" {
		t.Fatal()
	}
	if allowOriginValue(c, "https://evil.com") != "" {
		t.Fatal()
	}
}

func TestMethodOK(t *testing.T) {
	c := &corsConfig{allowMethods: []string{"GET", "POST"}}
	if !methodOK(c, "get") {
		t.Fatal()
	}
	if methodOK(c, "TRACE") {
		t.Fatal()
	}
}
