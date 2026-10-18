package ratelimit

import (
	"net"
	"net/http"
)

// ClientIP returns the client's IP address without the port.
func ClientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// Middleware wraps `next` and answers 429 Too Many Requests when the client
// identified by keyFunc has used up its quota.
func Middleware(l Limiter, keyFunc func(*http.Request) string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !l.Allow(keyFunc(r)) {
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}