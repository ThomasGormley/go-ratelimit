package limit

import (
	"context"
	"net/http"
	"sync"
	"time"
)

type FixedWindowLimiter struct {
	window    time.Duration
	threshold int

	requests   map[string]int
	requestsMu sync.Mutex

	windowOnce sync.Once
}

func FixedWindow() Middleware {
	limiter := FixedWindowLimiter{
		window:    time.Second * 5,
		threshold: 3,

		requests: make(map[string]int),
	}
	limiter.windowOnce.Do(func() {
		go limiter.handleWindow(context.Background())
	})

	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ip := getClientIP(r)
			if !limiter.allowed(ip) {
				w.WriteHeader(http.StatusTooManyRequests)
				return
			}
			next(w, r)
		}
	}
}

func (l *FixedWindowLimiter) handleWindow(ctx context.Context) {
	ticker := time.NewTicker(l.window)
	for {
		select {
		case <-ticker.C:
			go func() {
				l.requestsMu.Lock()
				defer l.requestsMu.Unlock()
				for k := range l.requests {
					delete(l.requests, k)
				}

			}()
		case <-ctx.Done():
			ticker.Stop()
			return
		}
	}
}

func (l *FixedWindowLimiter) allowed(ip string) bool {
	c := l.reqCount(ip)
	defer l.incrementRequestCount(ip)
	return c < l.threshold
}

func (l *FixedWindowLimiter) reqCount(ip string) int {
	l.requestsMu.Lock()
	defer l.requestsMu.Unlock()

	r, ok := l.requests[ip]

	if !ok {
		l.requests[ip] = 1
		return 1
	}

	return r
}

func (l *FixedWindowLimiter) incrementRequestCount(ip string) int {
	c := l.reqCount(ip)

	new := c + 1
	l.requests[ip] = new

	return new
}
