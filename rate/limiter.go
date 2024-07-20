package rate

type Limiter interface {
	// Limit checks if the rate limit has been reached.
	// It returns true if the limit has been reached, false otherwise.
	Limit(identifier string) bool
}
