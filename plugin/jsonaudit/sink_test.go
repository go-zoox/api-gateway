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
	"github.com/go-zoox/gormx"
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

func TestPrepare_JSONAuditOutputDatabaseMissingEngine(t *testing.T) {
	j := New()
	err := j.Prepare(defaults.Default(), &config.Config{
		JSONAudit: config.JSONAudit{
			Enable: true,
			Output: route.JSONAuditOutput{
				Provider: "database",
				Database: route.JSONAuditOutputDatabase{
					DSN: "db.sqlite",
				},
			},
		},
	})
	if err == nil || !strings.Contains(err.Error(), "output.database.engine") {
		t.Fatalf("expected error about output.database.engine, got %v", err)
	}
}

func TestPrepare_JSONAuditOutputDatabaseMissingDSN(t *testing.T) {
	j := New()
	err := j.Prepare(defaults.Default(), &config.Config{
		JSONAudit: config.JSONAudit{
			Enable: true,
			Output: route.JSONAuditOutput{
				Provider: "database",
				Database: route.JSONAuditOutputDatabase{
					Engine: "sqlite",
				},
			},
		},
	})
	if err == nil || !strings.Contains(err.Error(), "output.database.dsn") {
		t.Fatalf("expected error about output.database.dsn, got %v", err)
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

func TestEmitAuditLine_DatabaseSQLite(t *testing.T) {
	dbPath := t.TempDir() + string(os.PathSeparator) + "audit.sqlite"
	j := New()
	err := j.Prepare(defaults.Default(), &config.Config{
		JSONAudit: config.JSONAudit{
			Enable: true,
			Output: route.JSONAuditOutput{
				Provider: "database",
				Database: route.JSONAuditOutputDatabase{
					Engine: "sqlite",
					DSN:    dbPath,
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("prepare database sink failed: %v", err)
	}

	ctx := &zoox.Context{Logger: logger.New()}
	j.emitAuditLine(ctx, &j.globalConfig, []byte(`{"db":1}`))

	db := gormx.GetDB()
	if db == nil {
		t.Fatal("expected gormx db")
	}
	type rec struct {
		Payload string
	}
	var out rec
	if err := db.Raw("SELECT payload FROM json_audit_records ORDER BY id DESC LIMIT 1").Scan(&out).Error; err != nil {
		t.Fatalf("query migrated table failed: %v", err)
	}
	if out.Payload != `{"db":1}` {
		t.Fatalf("unexpected payload: %q", out.Payload)
	}
}

func TestPrepare_JSONAuditOutputDatabaseMultiConfigMismatch(t *testing.T) {
	j := New()
	err := j.Prepare(defaults.Default(), &config.Config{
		JSONAudit: config.JSONAudit{
			Enable: true,
			Output: route.JSONAuditOutput{
				Provider: "database",
				Database: route.JSONAuditOutputDatabase{
					Engine: "sqlite",
					DSN:    "a.sqlite",
				},
			},
		},
		Routes: []route.Route{
			{
				Path: "/api",
				JSONAudit: route.JSONAudit{
					Enable: true,
					Output: route.JSONAuditOutput{
						Provider: "database",
						Database: route.JSONAuditOutputDatabase{
							Engine: "sqlite",
							DSN:    "b.sqlite",
						},
					},
				},
			},
		},
	})
	if err == nil || !strings.Contains(err.Error(), "multiple database sinks") {
		t.Fatalf("expected mismatch error, got %v", err)
	}
}

func TestNormalizeJSONAuditDatabaseEngine(t *testing.T) {
	tests := map[string]string{
		"postgres":   "postgres",
		"PostgreSQL": "postgres",
		"pg":         "postgres",
		"mysql":      "mysql",
		"sqlite3":    "sqlite",
		"xxx":        "",
	}
	for in, want := range tests {
		if got := normalizeJSONAuditDatabaseEngine(in); got != want {
			t.Fatalf("engine %q => %q, want %q", in, got, want)
		}
	}
}

func TestResolveJSONAuditDatabaseConfig_PostgresURLWithoutEngine(t *testing.T) {
	engine, dsn, err := resolveJSONAuditDatabaseConfig(route.JSONAuditOutputDatabase{
		DSN: "postgres://u:p@127.0.0.1:5432/audit?sslmode=disable",
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if engine != "postgres" {
		t.Fatalf("engine=%s", engine)
	}
	if dsn != "postgres://u:p@127.0.0.1:5432/audit?sslmode=disable" {
		t.Fatalf("dsn=%s", dsn)
	}
}

func TestResolveJSONAuditDatabaseConfig_MySQLURLNormalize(t *testing.T) {
	engine, dsn, err := resolveJSONAuditDatabaseConfig(route.JSONAuditOutputDatabase{
		DSN: "mysql://root:secret@127.0.0.1:3306/apigw?charset=utf8mb4",
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if engine != "mysql" {
		t.Fatalf("engine=%s", engine)
	}
	if dsn != "root:secret@tcp(127.0.0.1:3306)/apigw?charset=utf8mb4" {
		t.Fatalf("dsn=%s", dsn)
	}
}

func TestResolveJSONAuditDatabaseConfig_SQLiteURLWithoutEngine(t *testing.T) {
	engine, dsn, err := resolveJSONAuditDatabaseConfig(route.JSONAuditOutputDatabase{
		DSN: "sqlite:///var/lib/api-gateway/json-audit.sqlite",
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if engine != "sqlite" {
		t.Fatalf("engine=%s", engine)
	}
	if dsn != "/var/lib/api-gateway/json-audit.sqlite" {
		t.Fatalf("dsn=%s", dsn)
	}
}

func TestResolveJSONAuditDatabaseConfig_KeepLegacyMySQLDSN(t *testing.T) {
	engine, dsn, err := resolveJSONAuditDatabaseConfig(route.JSONAuditOutputDatabase{
		Engine: "mysql",
		DSN:    "root:secret@tcp(127.0.0.1:3306)/apigw",
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if engine != "mysql" || dsn != "root:secret@tcp(127.0.0.1:3306)/apigw" {
		t.Fatalf("engine=%s dsn=%s", engine, dsn)
	}
}

func TestResolveJSONAuditDatabaseConfig_EngineMismatch(t *testing.T) {
	_, _, err := resolveJSONAuditDatabaseConfig(route.JSONAuditOutputDatabase{
		Engine: "postgres",
		DSN:    "mysql://root:secret@127.0.0.1:3306/apigw",
	})
	if err == nil || !strings.Contains(err.Error(), "does not match dsn scheme") {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestResolveJSONAuditDatabaseConfig_StructuredConfigPriorityHigherThanDSN(t *testing.T) {
	engine, dsn, err := resolveJSONAuditDatabaseConfig(route.JSONAuditOutputDatabase{
		Engine:   "postgres",
		Host:     "postgres",
		Port:     5432,
		Username: "postgres",
		Password: "postgres",
		DB:       "api-gateway",
		DSN:      "mysql://root:secret@127.0.0.1:3306/ignored",
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if engine != "postgres" {
		t.Fatalf("engine=%s", engine)
	}
	if !strings.Contains(dsn, "host=postgres") || !strings.Contains(dsn, "dbname=api-gateway") {
		t.Fatalf("dsn=%s", dsn)
	}
}

func TestResolveJSONAuditDatabaseConfig_PanicWhenDatabaseConfigMissing(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic when output.database config is missing")
		}
		msg, ok := r.(string)
		if !ok || !strings.Contains(msg, "requires dedicated output.database config") {
			t.Fatalf("unexpected panic: %v", r)
		}
	}()
	_, _, _ = resolveJSONAuditDatabaseConfig(route.JSONAuditOutputDatabase{})
}

func TestPrepare_JSONAuditOutputDatabaseMySQLAndPostgresValidationOnly(t *testing.T) {
	cfg := &config.Config{
		JSONAudit: config.JSONAudit{
			Enable: true,
			Output: route.JSONAuditOutput{
				Provider: "database",
				Database: route.JSONAuditOutputDatabase{
					Engine: "mysql",
					DSN:    "root:pass@tcp(localhost:3306)/db",
				},
			},
		},
	}
	// We don't connect to real MySQL/Postgres in unit tests. This ensures engine normalization accepts both.
	if got := normalizeJSONAuditDatabaseEngine(cfg.JSONAudit.Output.Database.Engine); got != "mysql" {
		t.Fatalf("mysql normalize got %q", got)
	}
	cfg.JSONAudit.Output.Database.Engine = "postgresql"
	if got := normalizeJSONAuditDatabaseEngine(cfg.JSONAudit.Output.Database.Engine); got != "postgres" {
		t.Fatalf("postgres normalize got %q", got)
	}
}

func TestEmitAuditLine_DatabaseFallbackIfNotInitialized(t *testing.T) {
	j := New()
	ctx := &zoox.Context{Logger: logger.New()}
	j.emitAuditLine(ctx, &route.JSONAudit{
		Enable: true,
		Output: route.JSONAuditOutput{
			Provider: "database",
			Database: route.JSONAuditOutputDatabase{
				Engine: "sqlite",
				DSN:    "x.sqlite",
			},
		},
	}, []byte(`{"ok":true}`))
}

func TestDatabaseSQLiteFileCreated(t *testing.T) {
	dbPath := t.TempDir() + string(os.PathSeparator) + "audit-check.sqlite"
	j := New()
	err := j.Prepare(defaults.Default(), &config.Config{
		JSONAudit: config.JSONAudit{
			Enable: true,
			Output: route.JSONAuditOutput{
				Provider: "database",
				Database: route.JSONAuditOutputDatabase{
					Engine: "sqlite",
					DSN:    dbPath,
				},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(dbPath); err != nil {
		t.Fatalf("sqlite db file not created: %v", err)
	}
}
