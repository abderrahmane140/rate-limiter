package main

import (
	"fmt"
	"sync"
	"time"

	"ratelimitdemo/ratelimit"
)

func main() {
	limiter := ratelimit.NewLimiter(3, 5*time.Second)

	fmt.Println("--- Part 1: two independent keys ---")
	for i := 1; i <= 4; i++ {
		fmt.Printf("alice request %d: allowed = %v\n", i, limiter.Allow("alice"))
	}
	for i := 1; i <= 4; i++ {
		fmt.Printf("bob   request %d: allowed = %v\n", i, limiter.Allow("bob"))
	}

	fmt.Println("\n--- Part 2: 50 concurrent requests for the same key ---")
	// Fresh limiter so Part 1's usage doesn't affect this demo.
	concurrentLimiter := ratelimit.NewLimiter(10, 5*time.Second)

	var wg sync.WaitGroup // lets us wait for all goroutines to finish
	var mu sync.Mutex     // guards our own `allowedCount` variable below
	allowedCount := 0

	for i := 0; i < 50; i++ {
		wg.Add(1) // tell the WaitGroup: "one more goroutine to wait for"
		go func() {
			defer wg.Done() // tell it: "this one is done" when we return
			if concurrentLimiter.Allow("shared-key") {
				mu.Lock()
				allowedCount++
				mu.Unlock()
			}
		}()
	}
	wg.Wait() // block here until all 50 goroutines have called Done()

	fmt.Printf("allowed %d out of 50 concurrent requests (limit was 10)\n", allowedCount)
}