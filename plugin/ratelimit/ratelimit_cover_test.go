package ratelimit

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-zoox/api-gateway/config"
	"github.com/go-zoox/api-gateway/core/route"
	"github.com/go-zoox/zoox"
	"github.com/go-zoox/zoox/defaults"
)

// boomStorage makes Allow fail to exercise OnRequest fail-open path.
type boomStorage struct{}

func (boomStorage) Allow(context.Context, string, int64, time.Duration) (bool, int64, time.Time, error) {
	return false, 0, time.Time{}, errors.New("allow failed")
}

func (boomStorage) Reset(context.Context, string) error { return nil }

func (boomStorage) Close() error { return nil }

func (boomStorage) LoadState(context.Context, string, string, interface{}) error {
	return errors.New("missing")
}

func (boomStorage) SaveState(context.Context, string, string, interface{}, time.Duration) error {
	return errors.New("save failed")
}

func TestAttachApplicationCache_NilApp(t *testing.T) {
	r := New()
	r.attachApplicationCache(nil)
	if r.limitStore != nil {
		t.Fatal("expected nil limitStore when app is nil")
	}
}

func TestRateLimit_NilStorageSkips(t *testing.T) {
	p := New()
	p.globalConfig = route.RateLimit{
		Enable: true, Limit: 5, Window: 60, KeyType: "ip",
	}
	p.attachApplicationCache(nil)
	req := httptest.NewRequest("GET", "/free", nil)
	ctx := createTestContext(req)
	if err := p.OnRequest(ctx, req); err != nil {
		t.Fatalf("OnRequest: %v", err)
	}
}

func TestRateLimit_AllowErrorFailOpen(t *testing.T) {
	p := New()
	p.globalConfig = route.RateLimit{
		Enable: true, Algorithm: "fixed-window", Limit: 1, Window: 60, KeyType: "ip",
	}
	p.limitStore = boomStorage{}
	req := httptest.NewRequest("GET", "/t", nil)
	ctx := createTestContext(req)
	if err := p.OnRequest(ctx, req); err != nil {
		t.Fatalf("expected fail-open nil error, got %v", err)
	}
}

func TestRateLimit_ExceededWithoutWriter(t *testing.T) {
	p := New()
	app := defaults.Default()
	cfg := &config.Config{
		RateLimit: route.RateLimit{
			Enable: true, Algorithm: "fixed-window",
			Limit: 1, Window: 60, KeyType: "ip",
		},
	}
	if err := p.Prepare(app, cfg); err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest("GET", "/only-one", nil)
	ctx := createTestContext(req)
	if err := p.OnRequest(ctx, req); err != nil {
		t.Fatalf("first: %v", err)
	}
	err := p.OnRequest(ctx, req)
	if err == nil {
		t.Fatal("expected rate limit error")
	}
}

func TestRateLimit_OnResponse_RetryAfter(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	c := req.Context()
	c = context.WithValue(c, "ratelimit:limit", int64(10))
	c = context.WithValue(c, "ratelimit:remaining", int64(0))
	c = context.WithValue(c, "ratelimit:reset", int64(1700000000))
	c = context.WithValue(c, "ratelimit:retryAfter", int64(42))
	req = req.WithContext(c)
	zctx := &zoox.Context{Request: req}
	res := &http.Response{Header: make(http.Header)}
	p := New()
	if err := p.OnResponse(zctx, res); err != nil {
		t.Fatal(err)
	}
	if res.Header.Get("Retry-After") != "42" {
		t.Fatalf("Retry-After = %q", res.Header.Get("Retry-After"))
	}
}

func TestAlgorithmFactory_AllBranches(t *testing.T) {
	f := &AlgorithmFactory{}
	if _, ok := f.NewAlgorithm("leaky-bucket").(*LeakyBucketAlgorithm); !ok {
		t.Fatal("leaky-bucket")
	}
	if _, ok := f.NewAlgorithm("fixed-window").(*FixedWindowAlgorithm); !ok {
		t.Fatal("fixed-window")
	}
	if _, ok := f.NewAlgorithm("weird-unknown").(*TokenBucketAlgorithm); !ok {
		t.Fatal("default should be token-bucket")
	}
}
