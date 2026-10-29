package cache

import (
	"context"
	"errors"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

func TestRedisOperations(t *testing.T) {
	addr := os.Getenv("TEST_REDIS_ADDR")
	if addr == "" {
		t.Skip("set TEST_REDIS_ADDR to run Redis integration tests")
	}
	opts := &redis.Options{Addr: addr}
	if addr[0] == '/' {
		opts.Network = "unix"
	}
	rdb := redis.NewClient(opts)
	t.Cleanup(func() { _ = rdb.Close() })
	c := &Client{rdb: rdb}
	ctx := context.Background()
	key := "test:blog:" + t.Name() + ":" + time.Now().Format("150405.000000000")
	t.Cleanup(func() { _ = rdb.Del(ctx, key, key+":missing").Err() })
	if _, err := c.Get(ctx, key); !errors.Is(err, ErrMiss) {
		t.Fatalf("missing key: %v", err)
	}
	if err := c.Set(ctx, key, "3", time.Minute); err != nil {
		t.Fatal(err)
	}
	values, err := c.PopBatch(ctx, []string{key, key + ":missing"})
	if err != nil || values[key] != "3" {
		t.Fatalf("partial batch = %v, %v", values, err)
	}
	var wg sync.WaitGroup
	limitKey := key + ":limit"
	t.Cleanup(func() { _ = rdb.Del(ctx, limitKey).Err() })
	var allowedCount atomic.Int64
	for range 20 {
		wg.Go(func() {
			ok, err := c.Allow(ctx, limitKey, 10, time.Minute)
			if err != nil {
				t.Error(err)
			}
			if ok {
				allowedCount.Add(1)
			}
		})
	}
	wg.Wait()
	if allowedCount.Load() != 10 {
		t.Fatalf("allowed requests = %d", allowedCount.Load())
	}
	if ttl := rdb.PTTL(ctx, limitKey).Val(); ttl <= 0 {
		t.Fatalf("limiter TTL = %v", ttl)
	}
}
