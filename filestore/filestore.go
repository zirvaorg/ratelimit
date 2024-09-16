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
	FilePath  string        // Path to the file where data is stored
	Rate      time.Duration // Rate limit duration
	Limit     int           // Number of allowed requests
	BlockTime time.Duration // Time to block after exceeding rate limit
}

// FileStore holds rate limit data in a file.
type FileStore struct {
	options Options
	mutex   sync.Mutex
}

// New creates a new FileStore with the given options.
func New(options Options) *FileStore {
	return &FileStore{
		options: options,
	}
}

// load reads the rate limit data from the file.
func (fs *FileStore) load() (map[string]*ratelimit.RateLimitInfo, error) {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	file, err := os.Open(fs.options.FilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]*ratelimit.RateLimitInfo), nil
		}
		return nil, err
	}
	defer file.Close()

	var data map[string]*ratelimit.RateLimitInfo
	if err := json.NewDecoder(file).Decode(&data); err != nil {
		return nil, err
	}

	return data, nil
}

// save writes the rate limit data to the file.
func (fs *FileStore) save(data map[string]*ratelimit.RateLimitInfo) error {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	file, err := os.Create(fs.options.FilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	if err := json.NewEncoder(file).Encode(&data); err != nil {
		return err
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
