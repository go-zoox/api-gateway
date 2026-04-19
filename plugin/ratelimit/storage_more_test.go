package ratelimit

import (
	"context"
	"errors"
	"testing"
	"time"

	zc "github.com/go-zoox/cache"
)

func TestStateCacheTTL(t *testing.T) {
	tests := []struct {
		name string
		win  time.Duration
		want time.Duration
	}{
		{"zero uses 24h", 0, 24 * time.Hour},
		{"small window clamps to 1h", time.Minute, time.Hour},
		{"10x window between 1h and 48h", 5 * time.Hour, 48 * time.Hour},
		{"large window capped at 48h", 10 * time.Hour, 48 * time.Hour},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stateCacheTTL(tt.win)
			if got != tt.want {
				t.Fatalf("stateCacheTTL(%v) = %v, want %v", tt.win, got, tt.want)
			}
		})
	}
}

// setFailCache delegates to an inner cache but fails Set when enabled.
type setFailCache struct {
	inner zc.Cache
	fail  bool
}

func (s *setFailCache) Get(key string, value interface{}) error {
	return s.inner.Get(key, value)
}

func (s *setFailCache) Set(key string, value interface{}, ttl ...time.Duration) error {
	if s.fail {
		return errors.New("injected set failure")
	}
	return s.inner.Set(key, value, ttl...)
}

func (s *setFailCache) Del(key string) error {
	return s.inner.Del(key)
}

func (s *setFailCache) Has(key string) bool {
	return s.inner.Has(key)
}

func (s *setFailCache) Clear() error {
	return s.inner.Clear()
}

func TestCacheStorage_Allow_SetErrorOnFirstWrite(t *testing.T) {
	inner := zc.New(&zc.Config{Engine: "memory"})
	wrap := &setFailCache{inner: inner, fail: true}
	s := newCacheStorage(wrap)
	ctx := context.Background()
	_, _, _, err := s.Allow(ctx, "k", 2, time.Second)
	if err == nil {
		t.Fatal("expected error from Set on first counter create")
	}
}

func TestCacheStorage_Allow_SetErrorOnUpdate(t *testing.T) {
	inner := zc.New(&zc.Config{Engine: "memory"})
	wrap := &setFailCache{inner: inner}
	s := newCacheStorage(wrap)
	ctx := context.Background()

	_, _, _, err := s.Allow(ctx, "upd", 3, time.Second)
	if err != nil {
		t.Fatalf("first Allow: %v", err)
	}

	wrap.fail = true
	_, _, _, err = s.Allow(ctx, "upd", 3, time.Second)
	if err == nil {
		t.Fatal("expected error from Set on counter update")
	}
}

func TestTokenBucket_SaveStateError(t *testing.T) {
	inner := zc.New(&zc.Config{Engine: "memory"})
	wrap := &setFailCache{inner: inner, fail: true}
	s := newCacheStorage(wrap)
	algo := &TokenBucketAlgorithm{}
	_, _, _, err := algo.Allow(context.Background(), s, "k", 5, time.Second, 0)
	if err == nil {
		t.Fatal("expected SaveState to fail")
	}
}

func TestLeakyBucket_SaveStateError(t *testing.T) {
	inner := zc.New(&zc.Config{Engine: "memory"})
	wrap := &setFailCache{inner: inner, fail: true}
	s := newCacheStorage(wrap)
	algo := &LeakyBucketAlgorithm{}
	_, _, _, err := algo.Allow(context.Background(), s, "k", 2, time.Second, 0)
	if err == nil {
		t.Fatal("expected SaveState to fail")
	}
}

func TestCacheStorage_Allow_SetErrorOnWindowReset(t *testing.T) {
	inner := zc.New(&zc.Config{Engine: "memory"})
	wrap := &setFailCache{inner: inner}
	s := newCacheStorage(wrap)
	ctx := context.Background()

	_, _, _, err := s.Allow(ctx, "rst", 2, 50*time.Millisecond)
	if err != nil {
		t.Fatalf("seed: %v", err)
	}
	for i := int64(0); i < 2; i++ {
		s.Allow(ctx, "rst", 2, 50*time.Millisecond)
	}

	time.Sleep(80 * time.Millisecond)

	wrap.fail = true
	_, _, _, err = s.Allow(ctx, "rst", 2, 50*time.Millisecond)
	if err == nil {
		t.Fatal("expected error from Set when resetting expired window")
	}
}
