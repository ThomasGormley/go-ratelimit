package rate

import (
	"fmt"
	"sync"
	"time"
)

func NewTokenBucketLimiter(bucketSize int, refreshInterval time.Duration) *TokenBucketRatelimiter {
	limiter := &TokenBucketRatelimiter{
		bucketSize:      bucketSize,
		refreshInterval: refreshInterval,
		requests:        make(map[string][]time.Time),
	}

	return limiter
}

type TokenBucketRatelimiter struct {
	bucketSize      int
	refreshInterval time.Duration

	requests    map[string][]time.Time
	requestsMu  sync.Mutex
	kickoffOnce sync.Once
}

func (rl *TokenBucketRatelimiter) Limit(ip string) bool {
	now := time.Now()
	// Check time since last request, and calculate any bucket refresh up to
	// maximum bucketSize
	rl.requestsMu.Lock()
	defer rl.requestsMu.Unlock()
	requests := lookup(rl.requests, ip)

	if len(rl.requests[ip]) > rl.bucketSize {
		// only keep the latest n (bucketSize) requests
		rl.requests[ip] = rl.requests[ip][len(rl.requests[ip])-rl.bucketSize:]
	}

	if len(requests) > 0 {
		lastRequest := requests[min(len(requests)-1, 0)]
		sinceLast := now.Sub(lastRequest)
		intervalsPassed := int(sinceLast / rl.refreshInterval)
		refreshTokens := intervalsPassed * 1
		fmt.Println("refreshTokens:", refreshTokens)
	}

	availableTokens := min(rl.bucketSize, len(requests))

	fmt.Println("availableTokens:", availableTokens)
	if availableTokens >= rl.bucketSize {
		return true
	}

	rl.requests[ip] = append(rl.requests[ip], now)
	return false
}

// func (rl *TokenBucketRatelimiter) kickoffRefreshSchedule(ctx context.Context) {
// 	ticker := time.NewTicker(rl.refreshInterval)
// 	for {
// 		select {
// 		case <-ticker.C:
// 			go func() {
// 				// refresh
// 				for ip := range rl.requests {
// 					rl.refreshBucket(ip)
// 				}
// 			}()
// 		case <-ctx.Done():
// 			ticker.Stop()
// 			return
// 		}
// 	}
// }

// func (rl *TokenBucketRatelimiter) refreshBucket(ip string) bool {
// 	r, ok := rl.requests[ip]
// 	if !ok || r == rl.bucketSize {
// 		return false
// 	}

// 	inc := r + 1
// 	invariant(inc <= rl.bucketSize, "Cannot increment greater than the bucket size")
// 	rl.requests[ip] = inc
// 	return true
// }

// func invariant(cond bool, msg string) {
// 	if !cond {
// 		panic(msg)
// 	}
// }
