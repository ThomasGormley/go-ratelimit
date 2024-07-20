package main

import (
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/thomasgormley/go-ratelimit/rhttp"
)

func handleLimited(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello, limited\n"))
}

func logRequest() rhttp.Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			defer func() {
				elapsed := time.Since(start)
				slog.Info("Request finished", "TIME", elapsed)
			}()
			next(w, r)
		}
	}
}

func wrap(handler http.HandlerFunc, middlewares ...rhttp.Middleware) http.HandlerFunc {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}

	return handler
}

func main() {

	mux := http.NewServeMux()
	// mux.HandleFunc("/limited", wrap(handleLimited, limit.TokenBucketRateLimiter(2)))
	mux.HandleFunc("/limited", wrap(handleLimited, logRequest(), rhttp.RateLimitFixedWindow()))
	mux.HandleFunc("/unlimited", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, unlimited\n"))
	})

	if err := http.ListenAndServe("localhost:8008", mux); err != nil {
		log.Fatalf("Error: %s", err)
	}
}
