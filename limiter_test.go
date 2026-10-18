package ratelimit

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestTokenBucket_BlocksAfterBurst(t *testing.T) {
	l := NewTokenBucket(3, time.Minute)

	for i := 1; i <= 3; i++ {
		if !l.Allow("alice") {
			t.Fatalf("request %d should be allowed", i)
		}
	}
	if l.Allow("alice") {
		t.Fatal("request 4 should be blocked")
	}
}

func TestTokenBucket_RefillsOverTime(t *testing.T) {
	l := NewTokenBucket(1, 100*time.Millisecond)

	l.Allow("alice")
	if l.Allow("alice") {
		t.Fatal("bucket should be empty")
	}

	time.Sleep(200 * time.Millisecond)
	if !l.Allow("alice") {
		t.Fatal("bucket should have refilled")
	}
}

func TestTokenBucket_KeysAreIndependent(t *testing.T) {
	l := NewTokenBucket(1, time.Minute)

	l.Allow("alice")
	if !l.Allow("bob") {
		t.Fatal("bob must not be affected by alice")
	}
}

func TestTokenBucket_CleanupRemovesIdleKeys(t *testing.T) {
	l := NewTokenBucket(1, 50*time.Millisecond)
	l.Allow("alice")

	time.Sleep(100 * time.Millisecond)
	l.Cleanup()

	if len(l.buckets) != 0 {
		t.Fatal("idle key should have been removed")
	}
}

func TestFixedWindow_ResetsAfterWindow(t *testing.T) {
	l := NewFixedWindow(2, 100*time.Millisecond)

	l.Allow("alice")
	l.Allow("alice")
	if l.Allow("alice") {
		t.Fatal("third request in the window should be blocked")
	}

	time.Sleep(200 * time.Millisecond)
	if !l.Allow("alice") {
		t.Fatal("a new window should allow requests again")
	}
}

func TestFixedWindow_SafeForConcurrentUse(t *testing.T) {
	l := NewFixedWindow(10, time.Minute)

	var allowed atomic.Int64
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if l.Allow("shared") {
				allowed.Add(1)
			}
		}()
	}
	wg.Wait()

	if allowed.Load() != 10 {
		t.Fatalf("allowed %d requests, want exactly 10", allowed.Load())
	}
}

func TestMiddleware_Returns429WhenLimitExceeded(t *testing.T) {
	l := NewFixedWindow(2, time.Minute)
	ok := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	h := Middleware(l, ClientIP, ok)

	want := []int{200, 200, 429}
	for i, code := range want {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = fmt.Sprintf("1.2.3.4:%d", 5000+i) // same IP, different port each time
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != code {
			t.Fatalf("request %d: got %d, want %d", i+1, rec.Code, code)
		}
	}
}