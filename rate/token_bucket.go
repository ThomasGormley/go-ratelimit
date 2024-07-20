package rate

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

func NewTokenBucketLimiter(bucketSize int, refreshInterval time.Duration) *TokenBucketRatelimiter {
	limiter := &TokenBucketRatelimiter{
		bucketSize:      bucketSize,
		refreshInterval: refreshInterval,
		requests:        make(map[string]int),
	}

	limiter.kickoffOnce.Do(func() {
		go limiter.kickoffRefreshSchedule(context.Background())
	})

	return limiter
}

type TokenBucketRatelimiter struct {
	bucketSize      int
	refreshInterval time.Duration

	requests    map[string]int
	requestsMu  sync.Mutex
	kickoffOnce sync.Once
}

func (rl *TokenBucketRatelimiter) Limit(ip string) bool {
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

	inc := r + 1
	invariant(inc <= rl.bucketSize, "Cannot increment greater than the bucket size")
	rl.requests[ip] = inc
	return true
}

func invariant(cond bool, msg string) {
	if !cond {
		panic(msg)
	}
}
