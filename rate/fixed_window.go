package rate

import (
	"context"
	"sync"
	"time"
)

func NewFixedWindowLimiter(window time.Duration, threshold int) *FixedWindowLimiter {
	limiter := &FixedWindowLimiter{
		window:    window,
		threshold: threshold,

		requests: make(map[string]int),
	}
	limiter.windowOnce.Do(func() {
		go limiter.handleWindow(context.Background())
	})

	return limiter
}

type FixedWindowLimiter struct {
	window    time.Duration
	threshold int

	requests   map[string]int
	requestsMu sync.Mutex

	windowOnce sync.Once
}

func (l *FixedWindowLimiter) Limit(ip string) bool {
	c := l.reqCount(ip)
	defer l.incrementReqCount(ip)
	return c > l.threshold
}

func (l *FixedWindowLimiter) handleWindow(ctx context.Context) {
	now := time.Now()
	// round down to the nearest window, then +1 window to get our end location
	nextWindow := now.Truncate(l.window).Add(l.window)
	timeUntilNextWindow := time.Until(nextWindow)
	untilNextWindowTicker := time.NewTicker(timeUntilNextWindow)

	select {
	case <-untilNextWindowTicker.C:
		// Handle the window we joined mid-way
		go l.clearRequests()
		// start a continuous window synced with clock
		ticker := time.NewTicker(l.window)
		for {
			select {
			case <-ticker.C:
				go l.clearRequests()
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
	case <-ctx.Done():
		untilNextWindowTicker.Stop()
		return
	}
}

func (l *FixedWindowLimiter) clearRequests() bool {
	l.requestsMu.Lock()
	defer l.requestsMu.Unlock()
	for k := range l.requests {
		delete(l.requests, k)
	}

	return true
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

func (l *FixedWindowLimiter) incrementReqCount(ip string) int {
	c := l.reqCount(ip)

	new := c + 1
	l.requests[ip] = new

	return new
}
