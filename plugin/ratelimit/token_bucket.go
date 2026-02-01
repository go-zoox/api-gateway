package ratelimit

import (
	"context"
	"time"
)

// TokenBucketAlgorithm implements token bucket rate limiting
type TokenBucketAlgorithm struct{}

// Allow checks if a request is allowed using token bucket algorithm
func (a *TokenBucketAlgorithm) Allow(ctx context.Context, storage Storage, key string, limit int64, window time.Duration, burst int64) (bool, int64, time.Time, error) {
	// Use burst if specified, otherwise use limit
	capacity := burst
	if capacity <= 0 {
		capacity = limit
	}

	// For token bucket, we use the storage with burst capacity
	// This allows burst traffic up to the capacity
	allowed, remaining, resetTime, err := storage.Allow(ctx, key, capacity, window)
	if err != nil {
		return false, 0, time.Time{}, err
	}

	return allowed, remaining, resetTime, nil
}
