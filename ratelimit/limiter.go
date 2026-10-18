package ratelimit


type Limiter interface {
	Allow(key string) bool
}