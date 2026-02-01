package ratelimit

import (
	"context"
	"sync"
	"time"
)

// tokenBucketState stores the state for a token bucket key
type tokenBucketState struct {
	tokens     float64   // current number of tokens in the bucket
	lastRefill time.Time // last time tokens were refilled
	mu         sync.Mutex
}

// TokenBucketAlgorithm implements token bucket rate limiting
type TokenBucketAlgorithm struct {
	mu     sync.RWMutex
	states map[string]*tokenBucketState
}

// getState gets or creates a state for the given key
func (a *TokenBucketAlgorithm) getState(key string, capacity int64) *tokenBucketState {
	a.mu.RLock()
	if a.states == nil {
		a.mu.RUnlock()
		a.mu.Lock()
		if a.states == nil {
			a.states = make(map[string]*tokenBucketState)
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
			state = &tokenBucketState{
				tokens:     float64(capacity), // start with full bucket
				lastRefill: time.Now(),
			}
			a.states[key] = state
		}
		a.mu.Unlock()
	}

	return state
}

// Allow checks if a request is allowed using token bucket algorithm
func (a *TokenBucketAlgorithm) Allow(ctx context.Context, storage Storage, key string, limit int64, window time.Duration, burst int64) (bool, int64, time.Time, error) {
	// Token bucket: tokens are added at a constant rate (limit/window)
	// The bucket can hold up to 'capacity' tokens (burst if specified, otherwise limit)
	// Each request consumes 1 token
	// If no tokens are available, the request is rejected

	// Use burst if specified, otherwise use limit
	capacity := burst
	if capacity <= 0 {
		capacity = limit
	}

	state := a.getState(key, capacity)
	state.mu.Lock()
	defer state.mu.Unlock()

	now := time.Now()

	// Calculate refill rate: limit tokens per window
	refillRate := float64(limit) / window.Seconds()

	// Calculate how many tokens to add since last refill
	elapsed := now.Sub(state.lastRefill).Seconds()
	tokensToAdd := refillRate * elapsed

	// Add tokens, but don't exceed capacity
	state.tokens += tokensToAdd
	if state.tokens > float64(capacity) {
		state.tokens = float64(capacity)
	}
	state.lastRefill = now

	// Check if we have at least 1 token
	if state.tokens < 1 {
		// No tokens available, reject request
		// Calculate when we'll have 1 token
		timeUntilToken := (1 - state.tokens) / refillRate
		resetTime := now.Add(time.Duration(timeUntilToken * float64(time.Second)))
		remaining := int64(0)
		return false, remaining, resetTime, nil
	}

	// Consume 1 token
	state.tokens -= 1

	// Calculate remaining tokens
	remaining := int64(state.tokens)
	if remaining < 0 {
		remaining = 0
	}

	// Calculate reset time (when bucket will be full)
	tokensNeeded := float64(capacity) - state.tokens
	timeUntilFull := tokensNeeded / refillRate
	resetTime := now.Add(time.Duration(timeUntilFull * float64(time.Second)))

	return true, remaining, resetTime, nil
}
