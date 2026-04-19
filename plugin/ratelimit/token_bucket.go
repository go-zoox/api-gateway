package ratelimit

import (
	"context"
	"time"
)

var tokenStripes stripes256

// tbState is persisted under Application.Cache() at ratelimit:tb:<client key>.
type tbState struct {
	Tokens     float64   `json:"tokens"`
	LastRefill time.Time `json:"last_refill"`
}

// TokenBucketAlgorithm implements token bucket rate limiting backed by Storage (Application.Cache).
type TokenBucketAlgorithm struct{}

// Allow checks if a request is allowed using token bucket logic with refill rate limit/window.
func (a *TokenBucketAlgorithm) Allow(ctx context.Context, storage Storage, key string, limit int64, window time.Duration, burst int64) (bool, int64, time.Time, error) {
	capacity := burst
	if capacity <= 0 {
		capacity = limit
	}

	unlock := tokenStripes.lock("tb:" + key)
	defer unlock()

	ttl := stateCacheTTL(window)

	var st tbState
	err := storage.LoadState(ctx, "tb", key, &st)
	now := time.Now()
	if err != nil {
		st = tbState{
			Tokens:     float64(capacity),
			LastRefill: now,
		}
	}

	refillRate := float64(limit) / window.Seconds()
	elapsed := now.Sub(st.LastRefill).Seconds()
	tokensToAdd := refillRate * elapsed

	st.Tokens += tokensToAdd
	if st.Tokens > float64(capacity) {
		st.Tokens = float64(capacity)
	}
	st.LastRefill = now

	if st.Tokens < 1 {
		timeUntilToken := (1 - st.Tokens) / refillRate
		resetTime := now.Add(time.Duration(timeUntilToken * float64(time.Second)))
		if err := storage.SaveState(ctx, "tb", key, &st, ttl); err != nil {
			return false, 0, time.Time{}, err
		}
		return false, 0, resetTime, nil
	}

	st.Tokens -= 1

	remaining := int64(st.Tokens)
	if remaining < 0 {
		remaining = 0
	}

	tokensNeeded := float64(capacity) - st.Tokens
	timeUntilFull := tokensNeeded / refillRate
	resetTime := now.Add(time.Duration(timeUntilFull * float64(time.Second)))

	if err := storage.SaveState(ctx, "tb", key, &st, ttl); err != nil {
		return false, 0, time.Time{}, err
	}
	return true, remaining, resetTime, nil
}
