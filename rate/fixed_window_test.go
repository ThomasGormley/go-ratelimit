package rate_test

import (
	"testing"
	"time"

	"github.com/thomasgormley/go-ratelimit/rate"
)

func TestFixedWindowLimiter_LimitsAndResets(t *testing.T) {
	t.Parallel()
	threshold, window := 2, time.Second*5
	limiter := rate.NewFixedWindowLimiter(window, threshold)
	client := "client:1"

	IsEqualBool(t, limiter.Limit(client), false)
	IsEqualBool(t, limiter.Limit(client), false)
	IsEqualBool(t, limiter.Limit(client), true)
	time.Sleep(window)
	IsEqualBool(t, limiter.Limit(client), false)
	IsEqualBool(t, limiter.Limit(client), false)
}
