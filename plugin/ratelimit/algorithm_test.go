package ratelimit

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"
)

var errMockStateMissing = errors.New("mock: state missing")

// mockStorage is a simple in-memory storage for testing
type mockStorage struct {
	counts map[string]int64
	resets map[string]time.Time
	states map[string][]byte // algorithm JSON: kind:key -> payload
}

func newMockStorage() *mockStorage {
	return &mockStorage{
		counts: make(map[string]int64),
		resets: make(map[string]time.Time),
		states: make(map[string][]byte),
	}
}

func mockStateKey(kind, key string) string {
	return kind + ":" + key
}

func (m *mockStorage) LoadState(ctx context.Context, kind, key string, dest interface{}) error {
	sk := mockStateKey(kind, key)
	raw, ok := m.states[sk]
	if !ok {
		return errMockStateMissing
	}
	return json.Unmarshal(raw, dest)
}

func (m *mockStorage) SaveState(ctx context.Context, kind, key string, src interface{}, ttl time.Duration) error {
	raw, err := json.Marshal(src)
	if err != nil {
		return err
	}
	if m.states == nil {
		m.states = make(map[string][]byte)
	}
	m.states[mockStateKey(kind, key)] = raw
	return nil
}

func (m *mockStorage) Allow(ctx context.Context, key string, limit int64, window time.Duration) (bool, int64, time.Time, error) {
	now := time.Now()
	resetTime, exists := m.resets[key]

	if !exists || now.After(resetTime) {
		m.counts[key] = 1
		m.resets[key] = now.Add(window)
		remaining := limit - 1
		if remaining < 0 {
			remaining = 0
		}
		return true, remaining, m.resets[key], nil
	}

	count := m.counts[key]
	if count >= limit {
		return false, 0, resetTime, nil
	}

	count++
	m.counts[key] = count
	remaining := limit - count
	if remaining < 0 {
		remaining = 0
	}
	return true, remaining, resetTime, nil
}

func (m *mockStorage) Reset(ctx context.Context, key string) error {
	delete(m.counts, key)
	delete(m.resets, key)
	delete(m.states, mockStateKey("tb", key))
	delete(m.states, mockStateKey("lb", key))
	return nil
}

func (m *mockStorage) Close() error {
	return nil
}

func TestFixedWindowAlgorithm(t *testing.T) {
	algorithm := &FixedWindowAlgorithm{}
	storage := newMockStorage()
	ctx := context.Background()

	key := "test-key"
	limit := int64(5)
	window := 1 * time.Second
	burst := int64(0)

	// First 5 requests should be allowed
	for i := int64(0); i < limit; i++ {
		allowed, remaining, _, err := algorithm.Allow(ctx, storage, key, limit, window, burst)
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
	}

	// 6th request should be denied
	allowed, remaining, _, err := algorithm.Allow(ctx, storage, key, limit, window, burst)
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

func TestTokenBucketAlgorithm(t *testing.T) {
	algorithm := &TokenBucketAlgorithm{}
	storage := newMockStorage()
	ctx := context.Background()

	key := "test-key-token"
	limit := int64(5)
	window := 1 * time.Second
	burst := int64(10) // Allow burst up to 10

	// Should allow up to burst capacity
	for i := int64(0); i < burst; i++ {
		allowed, remaining, _, err := algorithm.Allow(ctx, storage, key, limit, window, burst)
		if err != nil {
			t.Fatalf("Allow() error = %v", err)
		}
		if !allowed {
			t.Errorf("Allow() = false, want true for request %d (within burst)", i+1)
		}
		expectedRemaining := burst - (i + 1)
		if remaining != expectedRemaining {
			t.Errorf("Allow() remaining = %d, want %d", remaining, expectedRemaining)
		}
	}

	// Next request should be denied (burst exhausted)
	allowed, _, _, err := algorithm.Allow(ctx, storage, key, limit, window, burst)
	if err != nil {
		t.Fatalf("Allow() error = %v", err)
	}
	if allowed {
		t.Error("Allow() = true, want false (burst exceeded)")
	}
}

func TestLeakyBucketAlgorithm(t *testing.T) {
	algorithm := &LeakyBucketAlgorithm{}
	storage := newMockStorage()
	ctx := context.Background()

	key := "test-key-leaky"
	limit := int64(5)
	window := 1 * time.Second
	burst := int64(0) // Leaky bucket doesn't use burst

	// First 5 requests should be allowed
	for i := int64(0); i < limit; i++ {
		allowed, remaining, _, err := algorithm.Allow(ctx, storage, key, limit, window, burst)
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
	}

	// 6th request should be denied (no burst allowed)
	allowed, remaining, _, err := algorithm.Allow(ctx, storage, key, limit, window, burst)
	if err != nil {
		t.Fatalf("Allow() error = %v", err)
	}
	if allowed {
		t.Error("Allow() = true, want false (limit exceeded, no burst)")
	}
	if remaining != 0 {
		t.Errorf("Allow() remaining = %d, want 0", remaining)
	}
}

func TestTokenBucketAlgorithm_SharesStorageAcrossInstances(t *testing.T) {
	ctx := context.Background()
	storage := newTestCacheStorage()
	defer storage.Close()

	key := "same-client"
	limit := int64(5)
	window := time.Second
	burst := int64(2)

	a1 := &TokenBucketAlgorithm{}
	for i := int64(0); i < 2; i++ {
		allowed, _, _, err := a1.Allow(ctx, storage, key, limit, window, burst)
		if err != nil {
			t.Fatalf("Allow %d: %v", i, err)
		}
		if !allowed {
			t.Fatalf("request %d should be allowed", i)
		}
	}

	a2 := &TokenBucketAlgorithm{}
	allowed, _, _, err := a2.Allow(ctx, storage, key, limit, window, burst)
	if err != nil {
		t.Fatal(err)
	}
	if allowed {
		t.Fatal("third request should be denied (capacity 2 shared via cache)")
	}
}

func TestLeakyBucketAlgorithm_SharesStorageAcrossInstances(t *testing.T) {
	ctx := context.Background()
	storage := newTestCacheStorage()
	defer storage.Close()

	key := "leaky-client"
	limit := int64(2)
	window := time.Second

	a1 := &LeakyBucketAlgorithm{}
	for i := 0; i < 2; i++ {
		allowed, _, _, err := a1.Allow(ctx, storage, key, limit, window, 0)
		if err != nil || !allowed {
			t.Fatalf("request %d: allowed=%v err=%v", i, allowed, err)
		}
	}

	a2 := &LeakyBucketAlgorithm{}
	allowed, _, _, err := a2.Allow(ctx, storage, key, limit, window, 0)
	if err != nil {
		t.Fatal(err)
	}
	if allowed {
		t.Fatal("third request should be denied (shared leaky state)")
	}
}

func TestAlgorithmFactory(t *testing.T) {
	factory := &AlgorithmFactory{}

	tests := []struct {
		name          string
		algorithmType string
	}{
		{"Token bucket", "token-bucket"},
		{"Leaky bucket", "leaky-bucket"},
		{"Fixed window", "fixed-window"},
		{"Default", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			algorithm := factory.NewAlgorithm(tt.algorithmType)
			if algorithm == nil {
				t.Fatal("NewAlgorithm() returned nil")
			}
		})
	}
}
