package ratelimit

import (
	"context"
	"time"
)

// Algorithm defines the interface for rate limiting algorithms
type Algorithm interface {
	// Allow checks if a request is allowed
	// Returns:
	//   - allowed: whether the request is allowed
	//   - remaining: remaining requests/tokens
	//   - resetTime: when the rate limit will reset
	//   - err: any error that occurred
	Allow(ctx context.Context, storage Storage, key string, limit int64, window time.Duration, burst int64) (allowed bool, remaining int64, resetTime time.Time, err error)
}

// AlgorithmFactory creates algorithm instances
type AlgorithmFactory struct{}

// NewAlgorithm creates an algorithm instance based on algorithm type
func (f *AlgorithmFactory) NewAlgorithm(algorithmType string) Algorithm {
	switch algorithmType {
	case "token-bucket":
		return &TokenBucketAlgorithm{}
	case "leaky-bucket":
		return &LeakyBucketAlgorithm{}
	case "fixed-window", "":
		return &FixedWindowAlgorithm{}
	default:
		return &FixedWindowAlgorithm{} // default to fixed window
	}
}
