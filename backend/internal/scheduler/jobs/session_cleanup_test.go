package jobs

import (
	"context"
	"testing"
	"time"
)

type cleanupStub struct {
	cancel context.CancelFunc
	calls  int
}

func (s *cleanupStub) CleanupExpired(ctx context.Context) error {
	s.calls++
	if _, ok := ctx.Deadline(); !ok {
		panic("cleanup must have a deadline")
	}
	s.cancel()
	return ctx.Err()
}

func TestSessionCleanupRunsOnStartupAndStops(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	store := &cleanupStub{cancel: cancel}
	done := make(chan struct{})
	go func() {
		StartSessionCleanupJob(ctx, store, nil)
		close(done)
	}()
	select {
	case <-done:
		if store.calls != 1 {
			t.Fatalf("cleanup calls=%d", store.calls)
		}
	case <-time.After(time.Second):
		t.Fatal("cleanup did not stop on cancellation")
	}
}
