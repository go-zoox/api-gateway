package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/go-zoox/kv"
)

// RedisStorage implements Redis-based rate limit storage
type RedisStorage struct {
	cache kv.KV
}

// NewRedisStorage creates a new Redis storage
// cache can be kv.KV interface or kv.Config
func NewRedisStorage(cache interface{}) (*RedisStorage, error) {
	// Try to get kv.KV directly
	if kvCache, ok := cache.(kv.KV); ok {
		return &RedisStorage{
			cache: kvCache,
		}, nil
	}

	// Try to create from kv.Config
	if kvConfig, ok := cache.(kv.Config); ok {
		kvCache, err := kv.New(&kvConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create kv from config: %w", err)
		}
		return &RedisStorage{
			cache: kvCache,
		}, nil
	}

	return nil, fmt.Errorf("cache must be kv.KV or kv.Config, got %T", cache)
}

// Allow checks if a request is allowed using Redis
func (s *RedisStorage) Allow(ctx context.Context, key string, limit int64, window time.Duration) (bool, int64, time.Time, error) {
	now := time.Now()
	redisKey := fmt.Sprintf("ratelimit:%s", key)

	// Get current count
	var count int64
	if err := s.cache.Get(redisKey, &count); err != nil {
		// Key doesn't exist, create it with count = 1
		count = 1
		if err := s.cache.Set(redisKey, count, window); err != nil {
			return false, 0, time.Time{}, fmt.Errorf("failed to set rate limit key: %w", err)
		}
		resetTime := now.Add(window)
		remaining := limit - count
		if remaining < 0 {
			remaining = 0
		}
		return true, remaining, resetTime, nil
	}

	// Check if limit exceeded
	if count >= limit {
		// Estimate reset time based on window
		// Note: This is approximate since we don't have TTL info
		resetTime := now.Add(window)
		return false, 0, resetTime, nil
	}

	// Increment count
	count++
	if err := s.cache.Set(redisKey, count, window); err != nil {
		return false, 0, time.Time{}, fmt.Errorf("failed to update rate limit key: %w", err)
	}

	remaining := limit - count
	if remaining < 0 {
		remaining = 0
	}
	resetTime := now.Add(window)

	return true, remaining, resetTime, nil
}

// Reset resets the rate limit for a key
func (s *RedisStorage) Reset(ctx context.Context, key string) error {
	redisKey := fmt.Sprintf("ratelimit:%s", key)
	return s.cache.Delete(redisKey)
}

// Close closes the storage
func (s *RedisStorage) Close() error {
	// Redis connection is managed by the cache, nothing to close here
	return nil
}
