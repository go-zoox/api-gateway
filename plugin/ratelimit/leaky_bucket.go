package ratelimit

import (
	"context"
	"sync"
	"time"
)

// leakyBucketState stores the state for a leaky bucket key
type leakyBucketState struct {
	level      float64   // current water level in the bucket
	lastUpdate time.Time // last time the bucket was updated
	mu         sync.Mutex
}

// LeakyBucketAlgorithm implements leaky bucket rate limiting
type LeakyBucketAlgorithm struct {
	mu     sync.RWMutex
	states map[string]*leakyBucketState
}

// getState gets or creates a state for the given key
func (a *LeakyBucketAlgorithm) getState(key string) *leakyBucketState {
	a.mu.RLock()
	if a.states == nil {
		a.mu.RUnlock()
		a.mu.Lock()
		if a.states == nil {
			a.states = make(map[string]*leakyBucketState)
		}
		a.mu.Unlock()
		a.mu.RLock()
	}
	state, exists := a.states[key]
	a.mu.RUnlock()

	if !exists {
		a.mu.Lock()
		state, exists = a.states[key]
		if !exists {
			state = &leakyBucketState{
				level:      0,
				lastUpdate: time.Now(),
			}
			a.states[key] = state
		}
		a.mu.Unlock()
	}

	return state
}

// Allow checks if a request is allowed using leaky bucket algorithm
func (a *LeakyBucketAlgorithm) Allow(ctx context.Context, storage Storage, key string, limit int64, window time.Duration, burst int64) (bool, int64, time.Time, error) {
	// Leaky bucket: constant rate processing
	// The bucket leaks at a constant rate (limit/window)
	// Each request adds 1 to the bucket
	// If the bucket would overflow (level + 1 > limit), request is rejected

	state := a.getState(key)
	state.mu.Lock()
	defer state.mu.Unlock()

	now := time.Now()

	// Calculate leak rate: limit requests per window
	leakRate := float64(limit) / window.Seconds()

	// Calculate how much has leaked since last update
	elapsed := now.Sub(state.lastUpdate).Seconds()
	leaked := leakRate * elapsed

	// Update the water level (subtract leaked amount)
	state.level -= leaked
	if state.level < 0 {
		state.level = 0
	}
	state.lastUpdate = now

	// Check if adding 1 request would overflow the bucket
	// Use a small epsilon to handle floating point precision issues
	if state.level+1 > float64(limit)+0.0001 {
		// Bucket would overflow, reject request
		// Calculate when bucket will have space (when level drops enough to accept 1 more)
		timeUntilSpace := (state.level + 1 - float64(limit)) / leakRate
		resetTime := now.Add(time.Duration(timeUntilSpace * float64(time.Second)))
		remaining := int64(0)
		return false, remaining, resetTime, nil
	}

	// Add the request to the bucket
	state.level += 1

	// Calculate remaining capacity
	remaining := int64(float64(limit) - state.level)
	if remaining < 0 {
		remaining = 0
	}

	// Calculate reset time (when bucket will be empty)
	timeUntilEmpty := state.level / leakRate
	resetTime := now.Add(time.Duration(timeUntilEmpty * float64(time.Second)))

	return true, remaining, resetTime, nil
}
