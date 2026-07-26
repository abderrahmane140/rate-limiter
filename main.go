package main

import (
	"fmt"
	"time"
)

type Limiter struct {
	limit       int           // max requests allowed per window
	window      time.Duration //how long a window lasts. e.g. 10 * time.Second
	count       int           // how many requests we've counted in the current window
	windowStart time.Time     // when the current window began
}

func NewLimiter(limit int, window time.Duration) *Limiter {
	return &Limiter{
		limit:       limit,
		window:      window,
		count:       0,
		windowStart: time.Now(),
	}
}

func (l *Limiter) Allow() bool {
	now := time.Now()	

	if now.Sub(l.windowStart) >= l.window {
		l.windowStart = now
		l.count = 0
	}

	// Is there room left in this window?
	if l.count < l.limit {
		l.count++
		return true
	}
	return false
}

func main() {
	// Allow at most 3 requests per 5-second window
	limiter := NewLimiter(3, 5*time.Second)

	for i := 1; i <= 5; i++ {
		allowed := limiter.Allow()
		fmt.Printf("request %d: allowd = %v\n", i, allowed)
	}
}
