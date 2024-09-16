package memstore

import (
	"sync"
	"time"

	"github.com/zirvaorg/ratelimit"
)

// Options represents the configuration options for the MemStore.
type Options struct {
	Rate            time.Duration
	Limit           int
	BlockTime       time.Duration
	CleanupInterval time.Duration
}

// MemStore holds rate limit data in memory.
type MemStore struct {
	sync.Mutex
	options  Options
	requests map[string]*clientInfo
}

type clientInfo struct {
	remaining int
	resetTime time.Time
	blocked   bool
}

// New creates a new MemStore with the given options.
func New(options Options) *MemStore {
	if options.CleanupInterval == 0 {
		options.CleanupInterval = 30 * time.Minute
	}

	store := &MemStore{
		options:  options,
		requests: make(map[string]*clientInfo),
	}
	go store.cleanupExpiredEntries()
	return store
}

// Allow checks if the client with the given key is allowed to make a request.
func (ls *MemStore) Allow(key string) (bool, *ratelimit.RateLimitInfo) {
	ls.Lock()
	defer ls.Unlock()

	now := time.Now()
	client, exists := ls.requests[key]

	if !exists || now.After(client.resetTime) {
		client = &clientInfo{
			remaining: ls.options.Limit,
			resetTime: now.Add(ls.options.Rate),
			blocked:   false,
		}
		ls.requests[key] = client
	}

	if client.blocked {
		if now.Before(client.resetTime) {
			return false, &ratelimit.RateLimitInfo{
				Remaining: client.remaining,
				ResetTime: client.resetTime,
				Blocked:   client.blocked,
			}
		}
		client.blocked = false
		client.remaining = ls.options.Limit
		client.resetTime = now.Add(ls.options.Rate)
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

// cleanupExpiredEntries periodically removes expired entries from the store.
func (ls *MemStore) cleanupExpiredEntries() {
	for {
		time.Sleep(ls.options.CleanupInterval)
		ls.Lock()
		now := time.Now()
		for key, client := range ls.requests {
			if now.After(client.resetTime) {
				delete(ls.requests, key)
			}
		}
		ls.Unlock()
	}
}
