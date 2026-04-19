package route

import "testing"

func TestEffectiveJSONAuditProvider(t *testing.T) {
	if g := EffectiveJSONAuditProvider(JSONAuditOutput{}); g != "console" {
		t.Fatalf("empty: %q", g)
	}
	if g := EffectiveJSONAuditProvider(JSONAuditOutput{Provider: "  HTTP  "}); g != "http" {
		t.Fatalf("http: %q", g)
	}
	if g := EffectiveJSONAuditProvider(JSONAuditOutput{Provider: "file"}); g != "file" {
		t.Fatalf("file: %q", g)
	}
	if g := EffectiveJSONAuditProvider(JSONAuditOutput{Provider: "weird"}); g != "console" {
		t.Fatalf("unknown -> console: %q", g)
	}
}
