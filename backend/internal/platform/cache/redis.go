// Package cache provides redis cache
package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"blog-server/internal/platform/config"

	"github.com/redis/go-redis/v9"
)

type Client struct {
	rdb *redis.Client
}

func (c *Client) PopBatch(ctx context.Context, keys []string) (map[string]string, error) {
	pipe := c.rdb.Pipeline()
	cmds := make(map[string]*redis.StringCmd)

	for _, key := range keys {
		cmds[key] = pipe.GetDel(ctx, key)
	}
	if _, err := pipe.Exec(ctx); err != nil && !errors.Is(err, redis.Nil) {
		return nil, err
	}
	results := make(map[string]string)
	for key, cmd := range cmds {
		val, err := cmd.Result()
		if errors.Is(err, redis.Nil) {
			continue
		}
		if err != nil {
			return results, err
		}
		results[key] = val
	}
	return results, nil
}

func (c *Client) Scan(ctx context.Context, pattern string, cursor uint64, count int64) (keys []string, nextCursor uint64, err error) {
	return c.rdb.Scan(ctx, cursor, pattern, count).Result()
}

const (
	defaultTTL time.Duration = 5 * time.Minute
)

func (c *Client) Incr(ctx context.Context, key string) (int64, error) {
	return c.rdb.Incr(ctx, key).Result()
}

func (c *Client) IncrBy(ctx context.Context, key string, value int64) (int64, error) {
	return c.rdb.IncrBy(ctx, key, value).Result()
}

func (c *Client) Delete(ctx context.Context, key string) error {
	return c.rdb.Del(ctx, key).Err()
}

func (c *Client) Exists(ctx context.Context, key string) (bool, error) {
	count, err := c.rdb.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (c *Client) Get(ctx context.Context, key string) (string, error) {
	result, err := c.rdb.Get(ctx, key).Result()
	if err == nil {
		return result, nil
	}
	if errors.Is(err, redis.Nil) {
		return "", ErrMiss
	}
	return "", err
}

func (c *Client) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	if ttl == 0 {
		ttl = defaultTTL
	}
	if ttl < 0 {
		return c.rdb.Set(ctx, key, value, 0).Err()
	}
	return c.rdb.Set(ctx, key, value, ttl).Err()
}

// Allow applies a fixed window atomically. Expiry is set only on the first hit.
func (c *Client) Allow(ctx context.Context, key string, limit int64, window time.Duration) (bool, error) {
	count, err := c.rdb.Eval(ctx, `
local count = redis.call('INCR', KEYS[1])
if count == 1 then redis.call('PEXPIRE', KEYS[1], ARGV[1]) end
return count`, []string{key}, window.Milliseconds()).Int64()
	return count <= limit, err
}

func (c *Client) Ping(ctx context.Context) error {
	return c.rdb.Ping(ctx).Err()
}

func (c *Client) Close() error {
	return c.rdb.Close()
}

func NewCacheClient(cfg *config.Config) (*Client, error) {
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

	return &Client{rdb: rdb}, nil
}
