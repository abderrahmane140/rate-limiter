package ratelimit

import (
    "sync"
    "time"
)

type fixedWindowState struct {
    count       int
    windowStart time.Time
}

type FixedWindowLimiter struct {
    mu      sync.Mutex
    limit   int
    window  time.Duration
    entries map[string]*fixedWindowState
}

func NewFixedWindowLimiter(
    limit int,
    windowLen time.Duration,
) *FixedWindowLimiter {
    return &FixedWindowLimiter{
        limit:   limit,
        window:  windowLen,
        entries: make(map[string]*fixedWindowState),
    }
}

func (l *FixedWindowLimiter) Allow(key string) bool {
    l.mu.Lock()
    defer l.mu.Unlock()

    now := time.Now()

    s, exists := l.entries[key]
    if !exists {
        s = &fixedWindowState{windowStart: now}
        l.entries[key] = s
    }

    if now.Sub(s.windowStart) >= l.window {
        s.windowStart = now
        s.count = 0
    }

    if s.count < l.limit {
        s.count++
        return true
    }

    return false
}