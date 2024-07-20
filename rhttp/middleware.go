package rhttp

import (
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/thomasgormley/go-ratelimit/rate"
)

type Middleware func(next http.HandlerFunc) http.HandlerFunc

func RateLimitTokenBucket() Middleware {
	limiter := rate.NewTokenBucketLimiter(10, time.Millisecond*1000)

	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ip := getClientIP(r)
			if !limiter.Limit(ip) {
				w.WriteHeader(http.StatusTooManyRequests)
				return
			}
			next(w, r)
		}
	}
}

func RateLimitFixedWindow() Middleware {
	limiter := rate.NewFixedWindowLimiter(time.Second*10, 5)

	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ip := getClientIP(r)
			if limiter.Limit(ip) {
				w.WriteHeader(http.StatusTooManyRequests)
				return
			}
			next(w, r)
		}
	}
}

func RateLimitSlidingWindow() Middleware {
	limiter := rate.NewSlidingWindowLimiter(time.Second*10, 5)

	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ip := getClientIP(r)
			if limiter.Limit(ip) {
				w.WriteHeader(http.StatusTooManyRequests)
				return
			}
			next(w, r)
		}
	}
}

func getClientIP(r *http.Request) string {
	// Check the X-Forwarded-For header for the client IP
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		// X-Forwarded-For can contain a comma-separated list of IPs, take the first one
		ips := strings.Split(forwarded, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check the X-Real-IP header for the client IP
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Fallback to RemoteAddr if the headers are not set
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}
