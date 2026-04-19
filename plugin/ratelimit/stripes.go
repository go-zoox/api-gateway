package ratelimit

import (
	"hash/fnv"
	"sync"
)

// stripes256 hashes a logical key to one of 256 mutexes so the same client key
// serializes updates without pinning a mutex per unbounded key string.
type stripes256 struct {
	m [256]sync.Mutex
}

func (s *stripes256) lock(compositeKey string) func() {
	h := fnv.New32a()
	_, _ = h.Write([]byte(compositeKey))
	i := h.Sum32() % 256
	s.m[i].Lock()
	return func() { s.m[i].Unlock() }
}
