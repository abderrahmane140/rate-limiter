package main

import (
    "fmt"
    "time"

    "ratelimitdemo/ratelimit"
)

func runDemo(
    name string,
    limiter ratelimit.Limiter,
    requests int,
    delay time.Duration,
) {
    fmt.Println("---", name, "---")

    for i := 1; i <= requests; i++ {
        allowed := limiter.Allow("demo-user")
        fmt.Printf("Request %d: allowed = %v\n", i, allowed)
        time.Sleep(delay)
    }

    fmt.Println()
}

func main() {
    fixed := ratelimit.NewFixedWindowLimiter(
        3, 2*time.Second,
    )

    runDemo(
        "Fixed Window",
        fixed,
        6,
        300*time.Millisecond,
    )

    bucket := ratelimit.NewTokenBucketLimiter(
        3, 2*time.Second,
    )

    runDemo(
        "Token Bucket",
        bucket,
        6,
        300*time.Millisecond,
    )
}