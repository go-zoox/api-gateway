package ratelimit

import (
	"context"
	"sync"
	"time"
)

// MemoryStorage implements in-memory rate limit storage
type MemoryStorage struct {
	mu       sync.RWMutex
	store    map[string]*rateLimitEntry
	stopChan chan struct{}
}

type rateLimitEntry struct {
	count     int64
	resetTime time.Time
	mu        sync.Mutex
}

// NewMemoryStorage creates a new in-memory storage
func NewMemoryStorage() *MemoryStorage {
	storage := &MemoryStorage{
		store:    make(map[string]*rateLimitEntry),
		stopChan: make(chan struct{}),
	}

	// Start cleanup goroutine to remove expired entries
	go storage.cleanup()

	return storage
}

// Allow checks if a request is allowed
func (s *MemoryStorage) Allow(ctx context.Context, key string, limit int64, window time.Duration) (bool, int64, time.Time, error) {
	now := time.Now()

	s.mu.RLock()
	entry, exists := s.store[key]
	s.mu.RUnlock()

	if !exists {
		// Create new entry
		s.mu.Lock()
		entry, exists = s.store[key]
		if !exists {
			entry = &rateLimitEntry{
				count:     1,
				resetTime: now.Add(window),
			}
			s.store[key] = entry
			s.mu.Unlock()
			return true, limit - 1, entry.resetTime, nil
		}
		// Entry was created by another goroutine between RLock and Lock
		// We need to check the limit before allowing
		s.mu.Unlock()

		entry.mu.Lock()
		defer entry.mu.Unlock()

		// Check if window has expired
		if now.After(entry.resetTime) {
			entry.count = 1
			entry.resetTime = now.Add(window)
			return true, limit - 1, entry.resetTime, nil
		}

		// Check if limit exceeded
		if entry.count >= limit {
			return false, 0, entry.resetTime, nil
		}

		// Increment count
		entry.count++
		remaining := limit - entry.count
		if remaining < 0 {
			remaining = 0
		}
		return true, remaining, entry.resetTime, nil
	}

	entry.mu.Lock()
	defer entry.mu.Unlock()

	// Check if window has expired
	if now.After(entry.resetTime) {
		entry.count = 1
		entry.resetTime = now.Add(window)
		return true, limit - 1, entry.resetTime, nil
	}

	// Check if limit exceeded
	if entry.count >= limit {
		remaining := int64(0)
		if entry.count < limit {
			remaining = limit - entry.count
		}
		return false, remaining, entry.resetTime, nil
	}

	// Increment count
	entry.count++
	remaining := limit - entry.count
	if remaining < 0 {
		remaining = 0
	}

	return true, remaining, entry.resetTime, nil
}

// Reset resets the rate limit for a key
func (s *MemoryStorage) Reset(ctx context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.store, key)
	return nil
}

// Close closes the storage
func (s *MemoryStorage) Close() error {
	close(s.stopChan)
	return nil
}

// cleanup periodically removes expired entries
func (s *MemoryStorage) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			now := time.Now()
			s.mu.Lock()
			for key, entry := range s.store {
				entry.mu.Lock()
				if now.After(entry.resetTime) {
					delete(s.store, key)
				}
				entry.mu.Unlock()
			}
			s.mu.Unlock()
		}
	}
}
