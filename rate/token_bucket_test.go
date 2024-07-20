package rate_test

import (
	"testing"
	"time"

	"github.com/thomasgormley/go-ratelimit/rate"
)

func TestTokenBucketLimiter_LimitsAndRefillsBucket(t *testing.T) {
	bucket, interval := 1, time.Second*2
	limiter := rate.NewTokenBucketLimiter(bucket, interval)
	client := "client:1"

	IsEqualBool(t, limiter.Limit(client), false)
	IsEqualBool(t, limiter.Limit(client), true)
	time.Sleep(time.Second * 3)
	IsEqualBool(t, limiter.Limit(client), false)
}

// IsEqualBool fails test if got and want are not identical
func IsEqualBool(t *testing.T, got, want bool) {
	t.Helper()
	if got != want {
		t.Errorf("Assertion failed, got: %t, want: %t.", got, want)
	}
}
