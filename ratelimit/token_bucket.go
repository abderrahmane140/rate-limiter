package ratelimit

import (
    "sync"
    "time"
)

type tokenBucketState struct {
    tokens     float64
    lastRefill time.Time
}

type TokenBucketLimiter struct {
    mu         sync.Mutex
    capacity   float64
    refillRate float64
    buckets    map[string]*tokenBucketState
}

func NewTokenBucketLimiter(
    capacity int,
    per time.Duration,
) *TokenBucketLimiter {
    return &TokenBucketLimiter{
        capacity:   float64(capacity),
        refillRate: float64(capacity) / per.Seconds(),
        buckets:    make(map[string]*tokenBucketState),
    }
}

func (l *TokenBucketLimiter) Allow(key string) bool {
    l.mu.Lock()
    defer l.mu.Unlock()

    now := time.Now()

    b, exists := l.buckets[key]
    if !exists {
        b = &tokenBucketState{
            tokens:     l.capacity,
            lastRefill: now,
        }
        l.buckets[key] = b
    }

    elapsed := now.Sub(b.lastRefill).Seconds()
    b.tokens += elapsed * l.refillRate

    if b.tokens > l.capacity {
        b.tokens = l.capacity
    }

    b.lastRefill = now

    if b.tokens >= 1 {
        b.tokens--
        return true
    }

    return false
}