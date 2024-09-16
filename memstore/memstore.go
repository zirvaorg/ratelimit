package memstore

import (
	"sync"
	"time"

	"github.com/zirvaorg/ratelimit"
)

// Options represents the configuration options for the LocalStore.
type Options struct {
	Rate      time.Duration
	Limit     int
	BlockTime time.Duration
}

// LocalStore holds rate limit data in-memory.
type LocalStore struct {
	sync.Mutex
	options  Options
	requests map[string]*clientInfo
}

type clientInfo struct {
	remaining int
	resetTime time.Time
	blocked   bool
}

// New creates a new LocalStore with the given options.
func New(options Options) *LocalStore {
	return &LocalStore{
		options:  options,
		requests: make(map[string]*clientInfo),
	}
}

// Allow checks if a client with the given key can make a request.
func (ls *LocalStore) Allow(key string) (bool, *ratelimit.RateLimitInfo) {
	ls.Lock()
	defer ls.Unlock()

	now := time.Now()
	client, exists := ls.requests[key]

	if !exists || now.After(client.resetTime) {
		ls.requests[key] = &clientInfo{
			remaining: ls.options.Limit,
			resetTime: now.Add(ls.options.Rate),
			blocked:   false,
		}
		client = ls.requests[key]
	}

	if client.blocked && now.Before(client.resetTime) {
		return false, &ratelimit.RateLimitInfo{
			Remaining: client.remaining,
			ResetTime: client.resetTime,
			Blocked:   client.blocked,
		}
	}

	if now.After(client.resetTime) {
		client.remaining = ls.options.Limit
		client.resetTime = now.Add(ls.options.Rate)
		client.blocked = false
	}

	if client.remaining > 0 {
		client.remaining--
		return true, &ratelimit.RateLimitInfo{
			Remaining: client.remaining,
			ResetTime: client.resetTime,
			Blocked:   client.blocked,
		}
	}

	client.blocked = true
	client.resetTime = now.Add(ls.options.BlockTime)
	return false, &ratelimit.RateLimitInfo{
		Remaining: client.remaining,
		ResetTime: client.resetTime,
		Blocked:   client.blocked,
	}
}
