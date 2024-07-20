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

	l.requestsMu.Lock()
	defer l.requestsMu.Unlock()

	requestTimes := findOrInit(l.requests, ip)
	timesAfterWindow := pruneTimes(requestTimes, windowStart)

	if len(timesAfterWindow) >= l.threshold {
		return true
	}

	timesAfterWindow = append(timesAfterWindow, now)
	l.requests[ip] = timesAfterWindow
	return false
}

func findOrInit(m map[string][]time.Time, identifier string) []time.Time {
	t, ok := m[identifier]

	if !ok {
		t = []time.Time{}
		m[identifier] = t
	}

	return t
}

// pruneTimes removes all the time values from the given slice that are before the specified time.
// It returns a new slice containing only the time values that are after the specified time.
func pruneTimes(times []time.Time, after time.Time) []time.Time {
	timesAfter := []time.Time{}
	for _, time := range times {
		if time.After(after) {
			timesAfter = append(timesAfter, time)
		}
	}
	return timesAfter
}
