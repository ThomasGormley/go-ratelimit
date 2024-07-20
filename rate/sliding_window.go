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

	times := lookup(l.requestTimes, ip)
	timesAfterWindowStart := pruneTimes(times, windowStart)

	if len(timesAfterWindowStart) >= l.threshold {
		return true
	}

	timesAfterWindowStart = append(l.requestTimes[ip], now)
	l.requestTimes[ip] = timesAfterWindowStart
	return false
}
