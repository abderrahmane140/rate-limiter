// Package ratelimit implements a simple fixed-window rate limiter that
// tracks many independent keys and is safe to use from multiple
// goroutines at once.
package ratelimit

import (
	"sync"
	"time"
)

// window holds the counting state for ONE key.
// (lowercase name = unexported = only visible inside this package.
// We don't need callers outside this package to see or touch this.)
type window struct {
	count       int
	windowStart time.Time
}

// Limiter tracks rate limiting state for many keys at once.
type Limiter struct {
	mu      sync.Mutex // guards the map below
	limit   int
	window  time.Duration
	entries map[string]*window
}

// NewLimiter creates a Limiter allowing `limit` requests per `windowLen`,
// per key.
func NewLimiter(limit int, windowLen time.Duration) *Limiter {
	return &Limiter{
		limit:   limit,
		window:  windowLen,
		entries: make(map[string]*window), // maps must be created with make() before use
	}
}

// Allow reports whether one more request for `key` should be permitted
// right now. Safe to call concurrently from multiple goroutines.
func (l *Limiter) Allow(key string) bool {
	// Lock before touching the map, unlock when we're done -- this
	// ensures only one goroutine at a time can read or write l.entries.
	// `defer` schedules Unlock() to run right before Allow() returns,
	// no matter which return statement runs. This means we can't forget
	// to unlock, even if we add more return paths later.
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()

	w, exists := l.entries[key]
	if !exists {
		// First time we've seen this key: create its window.
		w = &window{windowStart: now}
		l.entries[key] = w
	}

	if now.Sub(w.windowStart) >= l.window {
		w.windowStart = now
		w.count = 0
	}

	if w.count < l.limit {
		w.count++
		return true
	}
	return false
}