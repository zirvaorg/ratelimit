package ratelimit

import (
	"net/http"
	"strconv"
	"time"
)

// RateLimitInfo contains information about rate limiting for a client.
type RateLimitInfo struct {
	Remaining int
	ResetTime time.Time
	Blocked   bool
}

// Store defines the interface for a rate limiting store.
type Store interface {
	Allow(key string) (bool, *RateLimitInfo)
}

// KeyFunc defines a function type for generating client keys.
type KeyFunc func(r *http.Request) string

// Middleware applies rate limiting to incoming requests.
func Middleware(store Store, next http.Handler, keyFunc KeyFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientKey := keyFunc(r)

		allowed, info := store.Allow(clientKey)

		// Add rate limiting headers
		w.Header().Set("X-RateLimit-Limit", strconv.Itoa(info.Remaining))
		w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(info.Remaining))
		w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(info.ResetTime.Unix(), 10))

		if allowed {
			next.ServeHTTP(w, r)
		} else {
			retryAfter := int(info.ResetTime.Sub(time.Now()).Seconds())
			w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
			http.Error(w, "Too many requests. Try again later.", http.StatusTooManyRequests)
		}
	})
}
