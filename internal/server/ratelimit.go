package server

import (
	"sync"
	"time"
)

// RateLimiter tracks request timestamps per IP per route (in-memory).
type RateLimiter struct {
	mu       sync.Mutex
	requests map[string]time.Time // key: "ip:path"
}

// NewRateLimiter creates a rate limiter with automatic cleanup of stale entries.
func NewRateLimiter() *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[string]time.Time),
	}
	go rl.cleanupLoop()
	return rl
}

// Allow checks if the IP is allowed to make a request on the given path.
func (rl *RateLimiter) Allow(ip, path string, window time.Duration) bool {
	key := ip + ":" + path
	rl.mu.Lock()
	defer rl.mu.Unlock()

	last, exists := rl.requests[key]
	if !exists || time.Since(last) >= window {
		rl.requests[key] = time.Now()
		return true
	}
	return false
}

func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		cutoff := time.Now().Add(-1 * time.Hour)
		for k, v := range rl.requests {
			if v.Before(cutoff) {
				delete(rl.requests, k)
			}
		}
		rl.mu.Unlock()
	}
}
