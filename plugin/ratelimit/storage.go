package ratelimit

import (
	"context"
	"fmt"
	"time"

	zc "github.com/go-zoox/cache"
)

var fixedWindowStripes stripes256

// Storage is the counter backend for rate-limit algorithms.
type Storage interface {
	Allow(ctx context.Context, key string, limit int64, window time.Duration) (allowed bool, remaining int64, resetTime time.Time, err error)
	Reset(ctx context.Context, key string) error
	Close() error

	// LoadState / SaveState persist token-bucket and leaky-bucket state in Application.Cache().
	// kind is "tb" or "lb"; key is the same logical client key used by extractors.
	LoadState(ctx context.Context, kind, key string, dest interface{}) error
	SaveState(ctx context.Context, kind, key string, src interface{}, ttl time.Duration) error
}

// counterEntry is stored in Application.Cache() (Redis JSON or memory KV under the hood).
type counterEntry struct {
	Count     int64     `json:"count"`
	ResetTime time.Time `json:"reset_time"`
}

// cacheStorage implements Storage using only zoox.Application.Cache() — no storage “type” switch.
type cacheStorage struct {
	c zc.Cache
}

func newCacheStorage(c zc.Cache) Storage {
	return &cacheStorage{c: c}
}

func algorithmStateKey(kind, key string) string {
	return fmt.Sprintf("ratelimit:%s:%s", kind, key)
}

func (s *cacheStorage) LoadState(ctx context.Context, kind, key string, dest interface{}) error {
	return s.c.Get(algorithmStateKey(kind, key), dest)
}

func (s *cacheStorage) SaveState(ctx context.Context, kind, key string, src interface{}, ttl time.Duration) error {
	return s.c.Set(algorithmStateKey(kind, key), src, ttl)
}

func (s *cacheStorage) Allow(ctx context.Context, key string, limit int64, window time.Duration) (bool, int64, time.Time, error) {
	unlock := fixedWindowStripes.lock("fw:" + key)
	defer unlock()

	now := time.Now()
	k := fmt.Sprintf("ratelimit:%s", key)

	var entry counterEntry
	if err := s.c.Get(k, &entry); err != nil {
		resetTime := now.Add(window)
		entry = counterEntry{Count: 1, ResetTime: resetTime}
		if err := s.c.Set(k, &entry, window); err != nil {
			return false, 0, time.Time{}, fmt.Errorf("set rate limit key: %w", err)
		}
		rem := limit - entry.Count
		if rem < 0 {
			rem = 0
		}
		return true, rem, resetTime, nil
	}

	if now.After(entry.ResetTime) {
		resetTime := now.Add(window)
		entry = counterEntry{Count: 1, ResetTime: resetTime}
		if err := s.c.Set(k, &entry, window); err != nil {
			return false, 0, time.Time{}, fmt.Errorf("reset rate limit key: %w", err)
		}
		rem := limit - entry.Count
		if rem < 0 {
			rem = 0
		}
		return true, rem, resetTime, nil
	}

	if entry.Count >= limit {
		return false, 0, entry.ResetTime, nil
	}

	entry.Count++
	ttl := time.Until(entry.ResetTime)
	if ttl <= 0 {
		rt := now.Add(window)
		entry.ResetTime = rt
		ttl = window
	}
	if err := s.c.Set(k, &entry, ttl); err != nil {
		return false, 0, time.Time{}, fmt.Errorf("update rate limit key: %w", err)
	}

	rem := limit - entry.Count
	if rem < 0 {
		rem = 0
	}
	return true, rem, entry.ResetTime, nil
}

func (s *cacheStorage) Reset(ctx context.Context, key string) error {
	_ = s.c.Del(fmt.Sprintf("ratelimit:%s", key))
	_ = s.c.Del(algorithmStateKey("tb", key))
	_ = s.c.Del(algorithmStateKey("lb", key))
	return nil
}

func (s *cacheStorage) Close() error {
	return nil
}

// stateCacheTTL is the Redis/memory key TTL for token- and leaky-bucket JSON blobs
// (refreshed on each request so active clients do not lose state).
func stateCacheTTL(window time.Duration) time.Duration {
	if window <= 0 {
		return 24 * time.Hour
	}
	t := 10 * window
	if t < time.Hour {
		t = time.Hour
	}
	if t > 48*time.Hour {
		t = 48 * time.Hour
	}
	return t
}
