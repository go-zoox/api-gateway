package jsonaudit

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/go-zoox/api-gateway/config"
	"github.com/go-zoox/api-gateway/core/route"
	"github.com/go-zoox/logger"
	"github.com/go-zoox/zoox"
	"github.com/go-zoox/zoox/defaults"
)

func TestPrepare_JSONAuditOutputFileMissingPath(t *testing.T) {
	j := New()
	err := j.Prepare(defaults.Default(), &config.Config{
		JSONAudit: config.JSONAudit{
			Enable: true,
			Output: route.JSONAuditOutput{
				Provider: "file",
			},
		},
	})
	if err == nil || !strings.Contains(err.Error(), "output.file.path") {
		t.Fatalf("expected error about output.file.path, got %v", err)
	}
}

func TestPrepare_JSONAuditOutputHTTPMissingURL(t *testing.T) {
	j := New()
	err := j.Prepare(defaults.Default(), &config.Config{
		JSONAudit: config.JSONAudit{
			Enable: true,
			Output: route.JSONAuditOutput{
				Provider: "http",
			},
		},
	})
	if err == nil || !strings.Contains(err.Error(), "output.http.url") {
		t.Fatalf("expected error about url, got %v", err)
	}
}

func TestEmitAuditLine_File(t *testing.T) {
	dir := t.TempDir()
	path := dir + string(os.PathSeparator) + "audit.ndjson"
	j := New()
	_ = j.Prepare(defaults.Default(), &config.Config{
		JSONAudit: config.JSONAudit{
			Enable: true,
			Output: route.JSONAuditOutput{
				Provider: "file",
				File:     route.JSONAuditOutputFile{Path: path},
			},
		},
	})
	ctx := &zoox.Context{Logger: logger.New()}
	cfg := &j.globalConfig
	j.emitAuditLine(ctx, cfg, []byte(`{"a":1}`))
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "{\"a\":1}\n" {
		t.Fatalf("got %q", b)
	}
}

func TestEmitAuditLine_HTTP(t *testing.T) {
	var gotMethod, gotCT, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotCT = r.Header.Get("Content-Type")
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)

	j := New()
	_ = j.Prepare(defaults.Default(), &config.Config{
		JSONAudit: config.JSONAudit{
			Enable: true,
			Output: route.JSONAuditOutput{
				Provider: "http",
				HTTP: route.JSONAuditHTTPOutput{
					URL:            srv.URL,
					TimeoutSeconds: 3,
				},
			},
		},
	})
	ctx := &zoox.Context{Logger: logger.New()}
	j.emitAuditLine(ctx, &j.globalConfig, []byte(`{"x":2}`))
	if gotMethod != http.MethodPost {
		t.Fatalf("method %s", gotMethod)
	}
	if !strings.Contains(gotCT, "application/json") {
		t.Fatalf("ct %q", gotCT)
	}
	if gotBody != `{"x":2}` {
		t.Fatalf("body %q", gotBody)
	}
}

func TestEmitAuditLine_HTTPFallbackOnError(t *testing.T) {
	j := New()
	_ = j.Prepare(defaults.Default(), &config.Config{
		JSONAudit: config.JSONAudit{
			Enable: true,
			Output: route.JSONAuditOutput{
				Provider: "http",
				HTTP: route.JSONAuditHTTPOutput{
					URL:            "http://127.0.0.1:1",
					TimeoutSeconds: 1,
				},
			},
		},
	})
	ctx := &zoox.Context{Logger: logger.New()}
	j.emitAuditLine(ctx, &j.globalConfig, []byte(`{"ok":true}`))
}
