package rate

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Middleware func(next http.HandlerFunc) http.HandlerFunc

type TokenBucketRatelimiter struct {
	bucketSize      int
	refreshInterval time.Duration

	requests    map[string]int
	requestsMu  sync.Mutex
	kickoffOnce sync.Once
}

func (rl *TokenBucketRatelimiter) kickoffRefreshSchedule(ctx context.Context) {
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
		case <-ctx.Done():
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

func (rl *TokenBucketRatelimiter) limit(ip string) bool {
	rl.requestsMu.Lock()
	defer rl.requestsMu.Unlock()

	remaining, ok := rl.requests[ip]
	if !ok {
		slog.Info("No entry found")
		rl.requests[ip] = rl.bucketSize
		remaining = rl.bucketSize
	}

	if limitReached := remaining <= 0; limitReached {
		return true
	} else {
		rl.requests[ip] = remaining - 1
		return false
	}
}

func TokenBucketRateLimiter(size int) Middleware {
	limiter := TokenBucketRatelimiter{
		bucketSize:      size,
		refreshInterval: time.Millisecond * 1000,
		requests:        make(map[string]int),
	}

	limiter.kickoffOnce.Do(func() {
		go limiter.kickoffRefreshSchedule(context.Background())
	})

	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ip := getClientIP(r)
			if !limiter.limit(ip) {
				w.WriteHeader(http.StatusTooManyRequests)
				return
			}
			next(w, r)
		}
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
