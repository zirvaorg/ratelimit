package filestore

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/zirvaorg/ratelimit"
)

// Options represents the configuration options for the FileStore.
type Options struct {
	FilePath        string
	Rate            time.Duration
	Limit           int
	BlockTime       time.Duration
	CleanupInterval time.Duration
}

// FileStore holds rate limit data in a file.
type FileStore struct {
	options Options
	mutex   sync.RWMutex
}

// New creates a new FileStore with the given options.
func New(options Options) *FileStore {
	if options.CleanupInterval == 0 {
		options.CleanupInterval = 30 * time.Minute
	}

	fs := &FileStore{
		options: options,
	}
	go fs.cleanupExpiredEntries()
	return fs
}

// load reads the rate limit data from the file.
func (fs *FileStore) load() (map[string]*ratelimit.RateLimitInfo, error) {
	fs.mutex.RLock()
	defer fs.mutex.RUnlock()

	file, err := os.Open(fs.options.FilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]*ratelimit.RateLimitInfo), nil
		}
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var data map[string]*ratelimit.RateLimitInfo
	if err := json.NewDecoder(file).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode file: %w", err)
	}

	return data, nil
}

// save writes the rate limit data to the file.
func (fs *FileStore) save(data map[string]*ratelimit.RateLimitInfo) error {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	file, err := os.Create(fs.options.FilePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	if err := json.NewEncoder(file).Encode(&data); err != nil {
		return fmt.Errorf("failed to encode data: %w", err)
	}

	return nil
}

// Allow checks if a client with the given key can make a request.
func (fs *FileStore) Allow(key string) (bool, *ratelimit.RateLimitInfo) {
	data, err := fs.load()
	if err != nil {
		fmt.Printf("Failed to load data: %v\n", err)
		return false, nil
	}

	info, exists := data[key]
	if !exists || time.Now().After(info.ResetTime) {
		info = &ratelimit.RateLimitInfo{
			Remaining: fs.options.Limit,
			ResetTime: time.Now().Add(fs.options.Rate),
			Blocked:   false,
		}
		data[key] = info
		fs.save(data)
	}

	if info.Blocked && time.Now().Before(info.ResetTime) {
		return false, info
	}

	if info.Remaining > 0 {
		info.Remaining--
		fs.save(data)
		return true, info
	}

	info.Blocked = true
	info.ResetTime = time.Now().Add(fs.options.BlockTime)
	fs.save(data)
	return false, info
}

// cleanupExpiredEntries periodically removes expired entries from the file.
func (fs *FileStore) cleanupExpiredEntries() {
	for {
		time.Sleep(fs.options.CleanupInterval)
		data, err := fs.load()
		if err != nil {
			fmt.Printf("Failed to load data: %v\n", err)
			continue
		}

		now := time.Now()
		for key, info := range data {
			if now.After(info.ResetTime) {
				delete(data, key)
			}
		}

		fs.save(data)
	}
}
