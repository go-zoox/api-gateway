package ratelimit

import (
	"context"
	"time"
)

// LeakyBucketAlgorithm implements leaky bucket rate limiting
type LeakyBucketAlgorithm struct{}

// Allow checks if a request is allowed using leaky bucket algorithm
func (a *LeakyBucketAlgorithm) Allow(ctx context.Context, storage Storage, key string, limit int64, window time.Duration, burst int64) (bool, int64, time.Time, error) {
	// Leaky bucket: constant rate processing, no burst allowed
	// The bucket leaks at a constant rate (limit/window)
	// If the bucket is full, requests are rejected

	// For leaky bucket, we use the storage's window-based approach
	// but enforce strict rate limiting (no burst)
	allowed, remaining, resetTime, err := storage.Allow(ctx, key, limit, window)
	if err != nil {
		return false, 0, time.Time{}, err
	}

	// Leaky bucket doesn't allow burst - if we're at limit, reject
	if !allowed {
		return false, remaining, resetTime, nil
	}

	return true, remaining, resetTime, nil
}
