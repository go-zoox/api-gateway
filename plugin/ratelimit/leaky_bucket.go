package ratelimit

import (
	"context"
	"time"
)

var leakyStripes stripes256

// lbState is persisted under Application.Cache() at ratelimit:lb:<client key>.
type lbState struct {
	Level      float64   `json:"level"`
	LastUpdate time.Time `json:"last_update"`
}

// LeakyBucketAlgorithm implements leaky bucket rate limiting backed by Storage (Application.Cache).
type LeakyBucketAlgorithm struct{}

// Allow checks if a request is allowed using leaky bucket logic.
func (a *LeakyBucketAlgorithm) Allow(ctx context.Context, storage Storage, key string, limit int64, window time.Duration, burst int64) (bool, int64, time.Time, error) {
	unlock := leakyStripes.lock("lb:" + key)
	defer unlock()

	ttl := stateCacheTTL(window)

	var st lbState
	err := storage.LoadState(ctx, "lb", key, &st)
	now := time.Now()
	if err != nil {
		st = lbState{
			Level:      0,
			LastUpdate: now,
		}
	}

	leakRate := float64(limit) / window.Seconds()

	elapsed := now.Sub(st.LastUpdate).Seconds()
	leaked := leakRate * elapsed

	st.Level -= leaked
	if st.Level < 0 {
		st.Level = 0
	}
	st.LastUpdate = now

	if st.Level+1 > float64(limit)+0.0001 {
		timeUntilSpace := (st.Level + 1 - float64(limit)) / leakRate
		resetTime := now.Add(time.Duration(timeUntilSpace * float64(time.Second)))
		if err := storage.SaveState(ctx, "lb", key, &st, ttl); err != nil {
			return false, 0, time.Time{}, err
		}
		return false, 0, resetTime, nil
	}

	st.Level += 1

	remaining := int64(float64(limit) - st.Level)
	if remaining < 0 {
		remaining = 0
	}

	timeUntilEmpty := st.Level / leakRate
	resetTime := now.Add(time.Duration(timeUntilEmpty * float64(time.Second)))

	if err := storage.SaveState(ctx, "lb", key, &st, ttl); err != nil {
		return false, 0, time.Time{}, err
	}

	return true, remaining, resetTime, nil
}
