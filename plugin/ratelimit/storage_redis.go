package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/go-zoox/kv"
)

// redisRateLimitEntry stores both count and resetTime to maintain fixed window
type redisRateLimitEntry struct {
	Count     int64     `json:"count"`
	ResetTime time.Time `json:"reset_time"`
}

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

	// Get current entry
	var entry redisRateLimitEntry
	if err := s.cache.Get(redisKey, &entry); err != nil {
		// Key doesn't exist, create it with count = 1
		resetTime := now.Add(window)
		entry = redisRateLimitEntry{
			Count:     1,
			ResetTime: resetTime,
		}
		if err := s.cache.Set(redisKey, entry, window); err != nil {
			return false, 0, time.Time{}, fmt.Errorf("failed to set rate limit key: %w", err)
		}
		remaining := limit - entry.Count
		if remaining < 0 {
			remaining = 0
		}
		return true, remaining, resetTime, nil
	}

	// Check if window has expired
	if now.After(entry.ResetTime) {
		// Window expired, reset count to 1
		resetTime := now.Add(window)
		entry = redisRateLimitEntry{
			Count:     1,
			ResetTime: resetTime,
		}
		if err := s.cache.Set(redisKey, entry, window); err != nil {
			return false, 0, time.Time{}, fmt.Errorf("failed to reset rate limit key: %w", err)
		}
		remaining := limit - entry.Count
		if remaining < 0 {
			remaining = 0
		}
		return true, remaining, resetTime, nil
	}

	// Check if limit exceeded
	if entry.Count >= limit {
		remaining := int64(0)
		return false, remaining, entry.ResetTime, nil
	}

	// Increment count - preserve existing TTL by calculating remaining time
	entry.Count++
	remainingTTL := time.Until(entry.ResetTime)
	if remainingTTL <= 0 {
		// TTL expired, reset the window
		resetTime := now.Add(window)
		entry.ResetTime = resetTime
		remainingTTL = window
	}
	// Update value with preserved TTL (remaining time until reset)
	if err := s.cache.Set(redisKey, entry, remainingTTL); err != nil {
		return false, 0, time.Time{}, fmt.Errorf("failed to update rate limit key: %w", err)
	}

	remaining := limit - entry.Count
	if remaining < 0 {
		remaining = 0
	}

	return true, remaining, entry.ResetTime, nil
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
