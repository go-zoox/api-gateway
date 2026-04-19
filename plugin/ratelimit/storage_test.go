package ratelimit

import (
	"context"
	"sync"
	"testing"
	"time"

	zc "github.com/go-zoox/cache"
)

func newTestCacheStorage() Storage {
	return newCacheStorage(zc.New(&zc.Config{Engine: "memory"}))
}

func TestCacheStorage_Allow(t *testing.T) {
	storage := newTestCacheStorage()
	defer storage.Close()

	ctx := context.Background()
	key := "test-key"
	limit := int64(5)
	window := 1 * time.Second

	// First 5 requests should be allowed
	for i := int64(0); i < limit; i++ {
		allowed, remaining, resetTime, err := storage.Allow(ctx, key, limit, window)
		if err != nil {
			t.Fatalf("Allow() error = %v", err)
		}
		if !allowed {
			t.Errorf("Allow() = false, want true for request %d", i+1)
		}
		expectedRemaining := limit - (i + 1)
		if remaining != expectedRemaining {
			t.Errorf("Allow() remaining = %d, want %d", remaining, expectedRemaining)
		}
		if resetTime.IsZero() {
			t.Error("Allow() resetTime is zero")
		}
	}

	// 6th request should be denied
	allowed, remaining, _, err := storage.Allow(ctx, key, limit, window)
	if err != nil {
		t.Fatalf("Allow() error = %v", err)
	}
	if allowed {
		t.Error("Allow() = true, want false (limit exceeded)")
	}
	if remaining != 0 {
		t.Errorf("Allow() remaining = %d, want 0", remaining)
	}
}

func TestCacheStorage_Allow_WindowExpiry(t *testing.T) {
	storage := newTestCacheStorage()
	defer storage.Close()

	ctx := context.Background()
	key := "test-key-expiry"
	limit := int64(3)
	window := 100 * time.Millisecond

	// Use up the limit
	for i := int64(0); i < limit; i++ {
		storage.Allow(ctx, key, limit, window)
	}

	// Wait for window to expire
	time.Sleep(150 * time.Millisecond)

	// Should be allowed again after window expiry
	allowed, remaining, _, err := storage.Allow(ctx, key, limit, window)
	if err != nil {
		t.Fatalf("Allow() error = %v", err)
	}
	if !allowed {
		t.Error("Allow() = false, want true (window expired)")
	}
	if remaining != limit-1 {
		t.Errorf("Allow() remaining = %d, want %d", remaining, limit-1)
	}
}

func TestCacheStorage_Reset(t *testing.T) {
	storage := newTestCacheStorage()
	defer storage.Close()

	ctx := context.Background()
	key := "test-key-reset"
	limit := int64(5)
	window := 1 * time.Second

	// Use up the limit
	for i := int64(0); i < limit; i++ {
		storage.Allow(ctx, key, limit, window)
	}

	// Reset
	if err := storage.Reset(ctx, key); err != nil {
		t.Fatalf("Reset() error = %v", err)
	}

	// Should be allowed again after reset
	allowed, remaining, _, err := storage.Allow(ctx, key, limit, window)
	if err != nil {
		t.Fatalf("Allow() error = %v", err)
	}
	if !allowed {
		t.Error("Allow() = false, want true (after reset)")
	}
	if remaining != limit-1 {
		t.Errorf("Allow() remaining = %d, want %d", remaining, limit-1)
	}
}

func TestCacheStorage_DifferentKeys(t *testing.T) {
	storage := newTestCacheStorage()
	defer storage.Close()

	ctx := context.Background()
	limit := int64(3)
	window := 1 * time.Second

	// Each key should have independent limits
	keys := []string{"key1", "key2", "key3"}
	for _, key := range keys {
		for i := int64(0); i < limit; i++ {
			allowed, _, _, err := storage.Allow(ctx, key, limit, window)
			if err != nil {
				t.Fatalf("Allow() error for key %s: %v", key, err)
			}
			if !allowed {
				t.Errorf("Allow() = false for key %s, want true", key)
			}
		}
	}
}

// TestCacheStorage_Allow_RaceCondition — concurrent first-touch; limit=2 ⇒ at most 2 allowed.
func TestCacheStorage_Allow_RaceCondition(t *testing.T) {
	storage := newTestCacheStorage()
	defer storage.Close()

	ctx := context.Background()
	key := "test-key-race"
	limit := int64(2)
	window := 1 * time.Second

	// Launch 3 goroutines simultaneously to race for creating the entry
	var wg sync.WaitGroup
	var allowedCount int64
	var mu sync.Mutex

	// Use a barrier to ensure all goroutines start at roughly the same time
	start := make(chan struct{})

	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start // Wait for signal to start
			allowed, _, _, err := storage.Allow(ctx, key, limit, window)
			if err != nil {
				t.Errorf("Allow() error = %v", err)
				return
			}
			mu.Lock()
			if allowed {
				allowedCount++
			}
			mu.Unlock()
		}()
	}

	// Signal all goroutines to start simultaneously
	close(start)
	wg.Wait()

	// Only 2 requests should be allowed, not 3
	if allowedCount != limit {
		t.Errorf("Allowed requests = %d, want %d (race condition not fixed)", allowedCount, limit)
	}
}
