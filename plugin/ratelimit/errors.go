package ratelimit

import "errors"

var (
	// ErrRateLimitExceeded is returned when rate limit is exceeded
	ErrRateLimitExceeded = errors.New("rate limit exceeded")
)
