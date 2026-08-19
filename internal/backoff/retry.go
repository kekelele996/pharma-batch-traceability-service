package backoff

import (
	"context"
	"fmt"
	"time"
)

// Do calls fn up to attempts times, sleeping backoff between failures. It
// stops early when ctx is cancelled and returns the last observed error after
// all attempts are exhausted.
func Do(ctx context.Context, attempts int, backoff time.Duration, fn func() error) error {
	if attempts < 1 {
		attempts = 1
	}
	var last error
	for i := 0; i < attempts; i++ {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("retry: aborted: %w", err)
		}
		if err := fn(); err == nil {
			return nil
		} else {
			last = err
		}
		if i == attempts-1 {
			break
		}
		if err := SleepCtx(ctx, backoff); err != nil {
			return err
		}
	}
	return fmt.Errorf("retry: exhausted after %d attempts: %w", attempts, last)
}

// SleepCtx sleeps for d, returning early when ctx is cancelled.
func SleepCtx(ctx context.Context, d time.Duration) error {
	if ctx.Err() != nil {
		return fmt.Errorf("retry: aborted: %w", ctx.Err())
	}
	select {
	case <-ctx.Done():
		return fmt.Errorf("retry: aborted: %w", ctx.Err())
	case <-time.After(d):
		return nil
	}
}
