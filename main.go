package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type TokenBucketRatelimiter struct {
	bucketSize      int
	refreshInterval time.Duration

	requests        map[string]int
	requestsMu      sync.Mutex
	tokenRefreshCtx context.Context
	kickoffOnce     sync.Once
}

func (rl *TokenBucketRatelimiter) kickoffRefreshSchedule() {
	ticker := time.NewTicker(rl.refreshInterval)
	for {
		select {
		case <-ticker.C:
			go func() {
				// refresh
				for ip := range rl.requests {
					rl.refreshBucket(ip)
				}
			}()
		case <-rl.tokenRefreshCtx.Done():
			ticker.Stop()
			return
		}
	}
}

func (rl *TokenBucketRatelimiter) refreshBucket(ip string) bool {
	r, ok := rl.requests[ip]
	if !ok || r == rl.bucketSize {
		return false
	}

	slog.Info("Incrementing request bucket for", "ip", ip)
	inc := r + 1
	invariant(inc <= rl.bucketSize, "Cannot increment greater than the bucket size")
	rl.requests[ip] = inc
	return true
}

func (rl *TokenBucketRatelimiter) allowRequest(ip string) bool {
	// kickoffRefreshScheduleOnce := sync.OnceFunc(rl.kickoffRefreshSchedule)
	// go kickoffRefreshScheduleOnce()

	rl.kickoffOnce.Do(func() {
		go rl.kickoffRefreshSchedule()
	})

	slog.Info("Evaluating for", "ip", ip)

	rl.requestsMu.Lock()
	defer rl.requestsMu.Unlock()

	remaining, ok := rl.requests[ip]
	if !ok {
		slog.Info("No entry found")
		rl.requests[ip] = rl.bucketSize
		remaining = rl.bucketSize

		fmt.Println("map:", rl.requests)
	}

	slog.Info("Remaining rquests", "r", remaining)
	if shouldAllow := remaining > 0; !shouldAllow {
		fmt.Println("Remaining return false")
		return false
	} else {
		// decrement requests
		fmt.Println("Decrementing request to ", remaining-1)
		rl.requests[ip] = remaining - 1
		return true
	}
}

func bucketLimiter(next http.HandlerFunc) http.HandlerFunc {
	limiter := TokenBucketRatelimiter{
		bucketSize:      2,
		refreshInterval: time.Millisecond * 1000,
		requests:        make(map[string]int),
		tokenRefreshCtx: context.Background(),
	}
	return func(w http.ResponseWriter, r *http.Request) {
		ip := getClientIP(r)
		if !limiter.allowRequest(ip) {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		next(w, r)
	}
}

func handleLimited(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello, limited\n"))
}

func main() {

	mux := http.NewServeMux()
	mux.HandleFunc("/limited", bucketLimiter(handleLimited))
	mux.HandleFunc("/unlimited", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, unlimited\n"))
	})

	if err := http.ListenAndServe("localhost:8008", mux); err != nil {
		log.Fatalf("Error: %s", err)
	}
}

func getClientIP(r *http.Request) string {
	// Check the X-Forwarded-For header for the client IP
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		// X-Forwarded-For can contain a comma-separated list of IPs, take the first one
		ips := strings.Split(forwarded, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check the X-Real-IP header for the client IP
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Fallback to RemoteAddr if the headers are not set
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

func invariant(cond bool, msg string) {
	if !cond {
		panic(msg)
	}
}
