package backoff

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestCtxDoAborts01(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := Do(ctx, 5, time.Millisecond, func() error { return errors.New("boom") })
	if err == nil || !strings.Contains(err.Error(), "aborted") {
		t.Fatalf("expected aborted error, got %v", err)
	}
}

func TestCtxSleepAborts02(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := SleepCtx(ctx, time.Millisecond)
	if err == nil || !strings.Contains(err.Error(), "aborted") {
		t.Fatalf("expected aborted error, got %v", err)
	}
}
