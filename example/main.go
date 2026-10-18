package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/abderrahmane140/rate-limiter"
)

func main() {
	// 5 requests at once, then 5 more per minute, for each client IP.
	limiter := ratelimit.NewTokenBucket(5, time.Minute)

	// Free memory used by clients that went away.
	go func() {
		for range time.Tick(time.Minute) {
			limiter.Cleanup()
		}
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "hello!")
	})

	handler := ratelimit.Middleware(limiter, ratelimit.ClientIP, mux)

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}