package ratelimit

import (
	"sync"
	"time"
)

// Limiter decides whether a request from `key` is allowed right now.
type Limiter interface {
	Allow(key string) bool
}

// ---------- Token bucket ----------

type bucket struct {
	tokens   float64
	lastSeen time.Time
}

// TokenBucket allows short bursts but enforces a steady average rate.
type TokenBucket struct {
	mu       sync.Mutex
	capacity float64 // max tokens a bucket can hold
	rate     float64 // tokens added per second
	buckets  map[string]*bucket
}

// NewTokenBucket allows `capacity` requests at once, then refills at
// `capacity` tokens per `per`. Example: NewTokenBucket(100, time.Minute).
func NewTokenBucket(capacity int, per time.Duration) *TokenBucket {
	return &TokenBucket{
		capacity: float64(capacity),
		rate:     float64(capacity) / per.Seconds(),
		buckets:  make(map[string]*bucket),
	}
}

// Allow reports whether one request for key may go through.
func (t *TokenBucket) Allow(key string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()

	b, ok := t.buckets[key]
	if !ok {
		b = &bucket{tokens: t.capacity, lastSeen: now} // new keys start full
		t.buckets[key] = b
	}

	// Refill based on how much time has passed since we last saw this key.
	elapsed := now.Sub(b.lastSeen).Seconds()
	b.tokens += elapsed * t.rate
	if b.tokens > t.capacity {
		b.tokens = t.capacity
	}
	b.lastSeen = now

	if b.tokens >= 1 {
		b.tokens--
		return true
	}
	return false
}

// Cleanup deletes keys that have been idle long enough for their bucket to be
// completely full again. Such keys behave exactly like new ones, so removing
// them never changes a decision; it just frees memory.
func (t *TokenBucket) Cleanup() {
	t.mu.Lock()
	defer t.mu.Unlock()

	timeToFill := time.Duration(t.capacity / t.rate * float64(time.Second))
	for key, b := range t.buckets {
		if time.Since(b.lastSeen) > timeToFill {
			delete(t.buckets, key)
		}
	}
}

// ---------- Fixed window ----------

type window struct {
	count int
	start time.Time
}

// FixedWindow allows `limit` requests per time window, per key.
type FixedWindow struct {
	mu      sync.Mutex
	limit   int
	size    time.Duration
	windows map[string]*window
}

// NewFixedWindow allows `limit` requests per `size`. Example:
// NewFixedWindow(10, time.Minute).
func NewFixedWindow(limit int, size time.Duration) *FixedWindow {
	return &FixedWindow{
		limit:   limit,
		size:    size,
		windows: make(map[string]*window),
	}
}

// Allow reports whether one request for key may go through.
func (f *FixedWindow) Allow(key string) bool {
	f.mu.Lock()
	defer f.mu.Unlock()

	now := time.Now()

	w, ok := f.windows[key]
	if !ok || now.Sub(w.start) >= f.size {
		w = &window{start: now} // first request, or the old window ended
		f.windows[key] = w
	}

	if w.count >= f.limit {
		return false
	}
	w.count++
	return true
}

// Cleanup deletes keys whose window has ended.
func (f *FixedWindow) Cleanup() {
	f.mu.Lock()
	defer f.mu.Unlock()

	for key, w := range f.windows {
		if time.Since(w.start) >= f.size {
			delete(f.windows, key)
		}
	}
}