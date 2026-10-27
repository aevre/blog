// Package cache provides redis cache
package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"blog-server/config"
	"blog-server/pkg/errx"

	"github.com/redis/go-redis/v9"
)

type client struct {
	rdb *redis.Client
}

// PopBatch implements [CacheClient].
func (c *client) PopBatch(ctx context.Context, keys []string) (map[string]string, error) {
	pipe := c.rdb.Pipeline()
	cmds := make(map[string]*redis.StringCmd)

	for _, key := range keys {
		cmds[key] = pipe.GetDel(ctx, key)
	}
	if _, err := pipe.Exec(ctx); err != nil {
		return nil, err
	}
	results := make(map[string]string)
	for key, cmd := range cmds {
		val, err := cmd.Result()
		if err != nil {
			continue
		}
		results[key] = val
	}
	return results, nil
}

// Scan implements [CacheClient].
func (c *client) Scan(ctx context.Context, pattern string, cursor uint64, count int64) (keys []string, nextCursor uint64, err error) {
	return c.rdb.Scan(ctx, cursor, pattern, count).Result()
}

const (
	defaultTTL time.Duration = 5 * time.Minute
)

// Incr implements [CacheClient].
func (c *client) Incr(ctx context.Context, key string) (int64, error) {
	return c.rdb.Incr(ctx, key).Result()
}

// IncrBy implements [CacheClient].
func (c *client) IncrBy(ctx context.Context, key string, value int64) (int64, error) {
	return c.rdb.IncrBy(ctx, key, value).Result()
}

func (c *client) Delete(ctx context.Context, key string) error {
	return c.rdb.Del(ctx, key).Err()
}

// Exists implements [CacheClient].
func (c *client) Exists(ctx context.Context, key string) (bool, error) {
	count, err := c.rdb.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Get implements [CacheClient].
func (c *client) Get(ctx context.Context, key string) (string, error) {
	result, err := c.rdb.Get(ctx, key).Result()
	if err == nil {
		return result, nil
	}
	if errors.Is(err, redis.Nil) {
		return "", errx.New(errx.CodeNotFound, err)
	}
	return "", err
}

// Set implements [CacheClient].
func (c *client) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	if ttl == 0 {
		ttl = defaultTTL
	}
	if ttl < 0 {
		return c.rdb.Set(ctx, key, value, 0).Err()
	}
	return c.rdb.Set(ctx, key, value, ttl).Err()
}

func (c *client) Close() error {
	return c.rdb.Close()
}

func NewCacheClient(cfg *config.Config) (CacheClient, error) {
	rcfg := cfg.Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:            fmt.Sprintf("%s:%d", rcfg.Host, rcfg.Port),
		Password:        rcfg.Password,
		DB:              rcfg.DB,
		PoolSize:        rcfg.PoolSize,
		MinIdleConns:    rcfg.MinIdleConns,
		DialTimeout:     rcfg.DialTimeout,
		ReadTimeout:     rcfg.ReadTimeout,
		WriteTimeout:    rcfg.WriteTimeout,
		PoolTimeout:     rcfg.PoolTimeout,
		ConnMaxIdleTime: rcfg.IdleTimeout,
	})

	pingTimeout := rcfg.DialTimeout
	if pingTimeout <= 0 {
		pingTimeout = 5 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), pingTimeout)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		_ = rdb.Close()
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	return &client{rdb: rdb}, nil
}
