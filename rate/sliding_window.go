package rate

import (
	"sync"
	"time"
)

func NewSlidingWindowLimiter(window time.Duration, threshold int) *SlidingWindowLimiter {
	limiter := &SlidingWindowLimiter{
		window:    window,
		threshold: threshold,

		requestTimes: make(map[string][]time.Time),
	}

	return limiter
}

type SlidingWindowLimiter struct {
	window    time.Duration
	threshold int

	requestTimes   map[string][]time.Time
	requestTimesMu sync.Mutex
}

func (l *SlidingWindowLimiter) Limit(ip string) bool {
	now := time.Now()
	windowStart := now.Add(-l.window)

	l.requestTimesMu.Lock()
	defer l.requestTimesMu.Unlock()

	times, ok := l.requestTimes[ip]

	if !ok {
		times = []time.Time{}
	}

	timesAfterWindowStart := []time.Time{}
	for _, t := range times {
		if t.After(windowStart) {
			timesAfterWindowStart = append(timesAfterWindowStart, t)
		}
	}

	l.requestTimes[ip] = timesAfterWindowStart

	if len(timesAfterWindowStart) >= l.threshold {
		return true
	}

	l.requestTimes[ip] = append(l.requestTimes[ip], now)
	return false
}
