# ratelimit
ratelimit is a simple rate limiting library for net/http handlers.
By default it can only store the rate limit in memory and file.

![goreport](https://goreportcard.com/badge/github.com/zirvaorg/ratelimit)
![license](https://badgen.net/github/license/zirvaorg/ratelimit)
[![pkg.go.dev](https://pkg.go.dev/badge/github.com/zirvaorg/ratelimit)](https://pkg.go.dev/github.com/zirvaorg/ratelimit)
![sourcegraph](https://sourcegraph.com/github.com/zirvaorg/ratelimit/-/badge.svg)

## Installation
```bash
go get github.com/zirvaorg/ratelimit
```

## MemStore Example
You can use `memstore` to keep data in memory. Below is an example usage.
```go
package main

import (
	"net/http"
	"time"

	"github.com/zirvaorg/ratelimit"
	"github.com/zirvaorg/ratelimit/memstore"
)

func keyFunc(r *http.Request) string {
	return r.RemoteAddr
}

func main() {
	store := memstore.New(memstore.Options{
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
```

## FileStore Example
You can use `filestore` to keep data in a file. Below is an example usage.
Filestore is not recommended for production use. But if you want to use it, you can use the following example.
```go
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
```