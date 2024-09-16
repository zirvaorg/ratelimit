package main

import (
	"net/http"
	"time"

	"github.com/zirvaorg/ratelimit"
	"github.com/zirvaorg/ratelimit/filestore"
)

func keyFunc(r *http.Request) string {
	return r.RemoteAddr
}

func main() {
	store := filestore.New(filestore.Options{
		FilePath:        "limit-db.json",
		Rate:            3 * time.Second,
		Limit:           10,
		BlockTime:       30 * time.Second,
		CleanupInterval: 30 * time.Minute,
	})

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, world!"))
	})

	http.Handle("/", ratelimit.Middleware(store, handler, keyFunc))

	http.ListenAndServe(":8080", nil)
}
