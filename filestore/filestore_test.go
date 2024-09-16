package filestore

import (
	"os"
	"testing"
	"time"
)

func TestFileStoreAllow(t *testing.T) {
	filePath := "limit-db.json"
	store := New(Options{
		FilePath:  filePath,
		Rate:      3 * time.Second,
		Limit:     5,
		BlockTime: 10 * time.Second,
	})

	key := "8.8.8.8"
	allowed, info := store.Allow(key)
	if !allowed {
		t.Fatalf("Expected to be allowed, but got blocked")
	}
	if info.Remaining != 4 {
		t.Fatalf("Expected remaining 4, but got %d", info.Remaining)
	}

	for i := 0; i < 6; i++ {
		store.Allow(key)
	}
	allowed, info = store.Allow(key)
	if allowed {
		t.Fatalf("Expected to be blocked, but got allowed")
	}
	if !info.Blocked {
		t.Fatalf("Expected to be blocked, but got not blocked")
	}

	os.Remove(filePath)
}

func TestFileStoreReset(t *testing.T) {
	filePath := "limit-db.json"
	store := New(Options{
		FilePath:  filePath,
		Rate:      3 * time.Second,
		Limit:     5,
		BlockTime: 10 * time.Second,
	})

	key := "8.8.8.8"
	for i := 0; i < 6; i++ {
		store.Allow(key)
	}

	time.Sleep(10 * time.Second)
	allowed, info := store.Allow(key)
	if !allowed {
		t.Fatalf("Expected to be allowed, but got blocked")
	}

	if info.Remaining != 4 {
		t.Fatalf("Expected remaining 4, but got %d", info.Remaining)
	}

	if info.Blocked {
		t.Fatalf("Expected to be not blocked, but got blocked")
	}

	os.Remove(filePath)
}
