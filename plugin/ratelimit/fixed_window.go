package ratelimit

import (
	"context"
	"time"
)

// FixedWindowAlgorithm implements fixed window rate limiting
type FixedWindowAlgorithm struct{}

// Allow checks if a request is allowed using fixed window algorithm
func (a *FixedWindowAlgorithm) Allow(ctx context.Context, storage Storage, key string, limit int64, window time.Duration, burst int64) (bool, int64, time.Time, error) {
	// Fixed window: simple count-based limiting within a time window
	return storage.Allow(ctx, key, limit, window)
}
