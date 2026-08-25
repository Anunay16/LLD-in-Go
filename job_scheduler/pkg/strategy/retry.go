package strategy

import (
	"math"
	"time"
)

// RetryPolicyStrategy determines whether a failed job should retry and the delay before retry.
type RetryPolicyStrategy interface {
	NextRetry(attempt int) (delay time.Duration, shouldRetry bool)
}

// NoRetryPolicy never retries on failure.
type NoRetryPolicy struct{}

func NewNoRetryPolicy() *NoRetryPolicy {
	return &NoRetryPolicy{}
}

func (p *NoRetryPolicy) NextRetry(attempt int) (time.Duration, bool) {
	return 0, false
}

// ExponentialBackoffPolicy calculates retry backoff exponentially with capped max duration.
type ExponentialBackoffPolicy struct {
	MaxRetries int
	BaseDelay  time.Duration
	MaxDelay   time.Duration
	Factor     float64
}

// NewExponentialBackoffPolicy creates an exponential backoff policy.
func NewExponentialBackoffPolicy(maxRetries int, baseDelay, maxDelay time.Duration) *ExponentialBackoffPolicy {
	return &ExponentialBackoffPolicy{
		MaxRetries: maxRetries,
		BaseDelay:  baseDelay,
		MaxDelay:   maxDelay,
		Factor:     2.0,
	}
}

func (p *ExponentialBackoffPolicy) NextRetry(attempt int) (time.Duration, bool) {
	if attempt > p.MaxRetries {
		return 0, false
	}

	// baseDelay * factor^(attempt - 1)
	multiplier := math.Pow(p.Factor, float64(attempt-1))
	delay := time.Duration(float64(p.BaseDelay) * multiplier)

	if delay > p.MaxDelay {
		delay = p.MaxDelay
	}

	return delay, true
}
