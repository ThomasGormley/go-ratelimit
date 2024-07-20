package main

import (
	"log"
	"net/http"

	"github.com/thomasgormley/go-ratelimit/limit"
)

func handleLimited(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello, limited\n"))
}

func wrap(handler http.HandlerFunc, middlewares ...limit.Middleware) http.HandlerFunc {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}

	return handler
}

func main() {

	mux := http.NewServeMux()
	// mux.HandleFunc("/limited", wrap(handleLimited, limit.TokenBucketRateLimiter(2)))
	mux.HandleFunc("/limited", wrap(handleLimited, limit.FixedWindow()))
	mux.HandleFunc("/unlimited", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, unlimited\n"))
	})

	if err := http.ListenAndServe("localhost:8008", mux); err != nil {
		log.Fatalf("Error: %s", err)
	}
}
