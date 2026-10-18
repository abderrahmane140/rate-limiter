# ratelimit

A small Go package for rate limiting. Two algorithms (token bucket and fixed
window), per-key limits, and an HTTP middleware. No dependencies.

## Install

```bash
go get github.com/yourusername/ratelimit
```

## Use it in your project

```go
limiter := ratelimit.NewTokenBucket(100, time.Minute) // 100 requests/minute per client

handler := ratelimit.Middleware(limiter, ratelimit.ClientIP, yourMux)
http.ListenAndServe(":8080", handler)
```

Clients over the limit get `429 Too Many Requests`.

### Without HTTP

```go
if limiter.Allow("user-123") {
    // do the work
} else {
    // too many requests
}
```

The string is the **key**. Every key has its own separate limit.

### Limit by something other than IP

```go
byAPIKey := func(r *http.Request) string { return r.Header.Get("X-API-Key") }
handler := ratelimit.Middleware(limiter, byAPIKey, yourMux)
```

### Important rules

- Create the limiter **once** at start-up and share it. Never create one per request.
- Call `limiter.Cleanup()` every minute or so in a long-running server, to free
  memory for clients that left (see `example/main.go`).
- Behind a reverse proxy, every request looks like it comes from the proxy's IP.
  Use a custom key function that reads a header your proxy sets.

## Choosing an algorithm

| Algorithm | Behavior |
|---|---|
| `NewTokenBucket(n, per)` | Allows bursts up to n, then a steady rate. Best default. |
| `NewFixedWindow(n, per)` | n requests per window. Simple, but can allow a double burst at the window edge. |

## Limitations

- Memory is per process. If you run several servers, each one has its own limits.
- Limits reset when the program restarts.

## Test

```bash
go test ./... -race
```

## License

MIT