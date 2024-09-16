package ratelimit

import (
	"net/http"
	"strconv"
	"time"
)

// RateLimitInfo contains information about rate limiting for a client.
type RateLimitInfo struct {
	Remaining int       // Number of remaining requests
	ResetTime time.Time // Time when the rate limit resets
	Blocked   bool      // Whether the client is blocked
}

// Store defines the interface for a rate limiting store.
type Store interface {
	Allow(key string) (bool, *RateLimitInfo)
}

// RateLimiter manages rate limiting using a given Store.
type RateLimiter struct {
	store Store
}

// NewRateLimiter creates a new RateLimiter with the given store.
func NewRateLimiter(store Store) *RateLimiter {
	return &RateLimiter{
		store: store,
	}
}

// Middleware applies rate limiting to incoming requests.
func (rl *RateLimiter) Middleware(next http.Handler, clientKey string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if clientKey == "" {
			// Use the client's IP address as the key not provided
			clientKey = r.RemoteAddr
		}

		allowed, info := rl.store.Allow(clientKey)

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
