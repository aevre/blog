// Package cache provide cache client interface
package cache

import (
	"context"
	"errors"
	"time"
)

// ErrMiss means the key does not exist; other errors indicate a store failure.
var ErrMiss = errors.New("cache miss")

type Store interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
}

type AtomicStore interface {
	Incr(ctx context.Context, key string) (int64, error)
	IncrBy(ctx context.Context, key string, value int64) (int64, error)
	PopBatch(ctx context.Context, keys []string) (map[string]string, error)
}

type PatternScanner interface {
	Scan(ctx context.Context, pattern string, cursor uint64, count int64) (keys []string, nextCursor uint64, err error)
}
