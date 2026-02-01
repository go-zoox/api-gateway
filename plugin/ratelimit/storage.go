package ratelimit

import (
	"context"
	"time"
)

// Storage defines the interface for rate limit storage
type Storage interface {
	// Allow checks if a request is allowed and updates the rate limit state
	// Returns:
	//   - allowed: whether the request is allowed
	//   - remaining: remaining requests in the current window
	//   - resetTime: when the rate limit will reset
	//   - err: any error that occurred
	Allow(ctx context.Context, key string, limit int64, window time.Duration) (allowed bool, remaining int64, resetTime time.Time, err error)

	// Reset resets the rate limit for a key
	Reset(ctx context.Context, key string) error

	// Close closes the storage and releases resources
	Close() error
}

// StorageFactory creates storage instances
type StorageFactory struct{}

// NewStorage creates a storage instance based on storage type
// cache can be kv.KV interface or kv.Config
func (f *StorageFactory) NewStorage(storageType string, cache interface{}) (Storage, error) {
	switch storageType {
	case "redis":
		return NewRedisStorage(cache)
	case "memory", "":
		return NewMemoryStorage(), nil
	default:
		return NewMemoryStorage(), nil // default to memory
	}
}
