package api

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"time"
)

// RateLimiter enforces a minimum delay between requests and provides retry backoff.
type RateLimiter struct {
	minDelay   time.Duration
	maxBackoff time.Duration
	lastSent   time.Time
	mu         chan struct{}
}

func NewRateLimiter(minDelay time.Duration, maxBackoff time.Duration) *RateLimiter {
	return &RateLimiter{
		minDelay:   minDelay,
		maxBackoff: maxBackoff,
		mu:         make(chan struct{}, 1),
	}
}

// Wait blocks until it is safe to send the next request.
func (r *RateLimiter) Wait(ctx context.Context) error {
	r.mu <- struct{}{}
	defer func() { <-r.mu }()

	elapsed := time.Since(r.lastSent)
	if elapsed < r.minDelay {
		sleep := r.minDelay - elapsed
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(sleep):
		}
	}
	r.lastSent = time.Now()
	return nil
}

// RetryWithBackoff executes fn with exponential backoff on transient failures.
func (r *RateLimiter) RetryWithBackoff(ctx context.Context, fn func() error) error {
	var lastErr error
	for attempt := 0; attempt < 5; attempt++ {
		if attempt > 0 {
			delay := r.backoffDelay(attempt)
			slog.Warn("retrying after error", "attempt", attempt, "delay", delay, "error", lastErr)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
		}
		if err := fn(); err == nil {
			return nil
		} else {
			lastErr = err
			if isTerminalError(err) {
				return err
			}
		}
	}
	return fmt.Errorf("exhausted retries (%d): %w", 5, lastErr)
}

func (r *RateLimiter) backoffDelay(attempt int) time.Duration {
	base := time.Duration(1<<attempt) * time.Second
	if base > r.maxBackoff {
		base = r.maxBackoff
	}
	// Add +/- 25% jitter
	jitter := time.Duration(rand.Int63n(int64(base)/2)) - time.Duration(int64(base)/4)
	return base + jitter
}

func isTerminalError(err error) bool {
	// 404s and context cancellation are terminal
	if err == context.Canceled || err == context.DeadlineExceeded {
		return true
	}
	return false
}
