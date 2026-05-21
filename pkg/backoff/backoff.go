package backoff

import (
	"math"
	"math/rand"
	"time"
)

// Delay returns a backoff duration for the given attempt (1-based) with full jitter.
// Formula: random uniform in [0, min(max, base * 2^(attempt-1))].
func Delay(attempt int, base, max time.Duration) time.Duration {
	if attempt < 1 {
		attempt = 1
	}
	exp := float64(base) * math.Pow(2, float64(attempt-1))
	if exp > float64(max) {
		exp = float64(max)
	}
	if exp <= 0 {
		return 0
	}
	jitter := rand.Float64() * exp
	return time.Duration(jitter)
}
