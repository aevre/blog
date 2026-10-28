package service

import (
	"context"
	"errors"
	"testing"
)

type atomicStoreStub struct {
	values map[string]int64
	fail   bool
}

func (s *atomicStoreStub) Incr(context.Context, string) (int64, error) { return 0, nil }
func (s *atomicStoreStub) IncrBy(_ context.Context, key string, value int64) (int64, error) {
	if s.fail {
		return 0, errors.New("redis unavailable")
	}
	s.values[key] += value
	return s.values[key], nil
}
func (s *atomicStoreStub) PopBatch(context.Context, []string) (map[string]string, error) {
	return nil, nil
}

func TestRestoreViewCounts(t *testing.T) {
	store := &atomicStoreStub{values: map[string]int64{"views:1": 2}}
	errs := restoreViewCounts(context.Background(), store, map[uint]int64{1: 3}, map[uint]string{1: "views:1"})
	if len(errs) != 0 {
		t.Fatalf("restore errors: %v", errs)
	}
	if got := store.values["views:1"]; got != 5 {
		t.Fatalf("restored count = %d, want 5", got)
	}
}
