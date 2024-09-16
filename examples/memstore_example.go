package main

import (
	"net/http"
	"time"

	"github.com/zirvaorg/ratelimit"
	"github.com/zirvaorg/ratelimit/memstore"
)

func main() {
	options := memstore.Options{
		Rate:      3 * time.Second,
		Limit:     10,
		BlockTime: 30 * time.Second,
	}

	store := memstore.New(options)
	rateLimiter := ratelimit.NewRateLimiter(store)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, world!"))
	})

	// since the clientKey value is empty, the ip address with `r.RemoteAddr` is used for default.
	// token or other identifying info can be added to this parameter to rate limit based on that.
	http.Handle("/", rateLimiter.Middleware(handler, ""))

	http.ListenAndServe(":8080", nil)
}
