package rate

import (
	"sync"
	"time"
)

func NewFixedWindowLimiter(window time.Duration, threshold int) *FixedWindowLimiter {
	limiter := &FixedWindowLimiter{
		window:    window,
		threshold: threshold,

		requests: make(map[string][]time.Time),
	}

	return limiter
}

type FixedWindowLimiter struct {
	window    time.Duration
	threshold int

	requests   map[string][]time.Time
	requestsMu sync.Mutex
}

func (l *FixedWindowLimiter) Limit(ip string) bool {
	now := time.Now()
	windowStart := now.Truncate(l.window)

	requestTimes, ok := l.requests[ip]

	if !ok {
		requestTimes = []time.Time{}
	}

	timesAfterWindow := []time.Time{}

	for _, time := range requestTimes {
		if time.After(windowStart) {
			timesAfterWindow = append(timesAfterWindow, time)
		}
	}

	if len(timesAfterWindow) >= l.threshold {
		return true
	}

	timesAfterWindow = append(timesAfterWindow, now)
	l.requests[ip] = timesAfterWindow
	return false
}
