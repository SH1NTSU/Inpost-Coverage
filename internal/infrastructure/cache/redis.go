package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Redis struct {
	rdb *redis.Client
	ttl time.Duration
}

// RedisConfig captures every way you might want to point at a Redis server.
// Managed Redis (Railway / Upstash / Render) hands you a URL like
// `rediss://default:pass@host:6380` — pass it via URL. Local development
// without a password just uses Addr.
type RedisConfig struct {
	URL      string
	Addr     string
	Username string
	Password string
	DB       int
}

func NewRedis(cfg RedisConfig, ttl time.Duration) (*Redis, error) {
	var opts *redis.Options
	if cfg.URL != "" {
		parsed, err := redis.ParseURL(cfg.URL)
		if err != nil {
			return nil, fmt.Errorf("parse REDIS_URL: %w", err)
		}
		opts = parsed
	} else {
		opts = &redis.Options{
			Addr:     cfg.Addr,
			Username: cfg.Username,
			Password: cfg.Password,
			DB:       cfg.DB,
		}
	}
	opts.DialTimeout = 2 * time.Second
	opts.ReadTimeout = 500 * time.Millisecond
	opts.WriteTimeout = 500 * time.Millisecond

	rdb := redis.NewClient(opts)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, err
	}
	return &Redis{rdb: rdb, ttl: ttl}, nil
}

func (r *Redis) Get(ctx context.Context, key string, dst any) (bool, error) {
	buf, err := r.rdb.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil
		}
		return false, err
	}
	if err := unmarshal(buf, dst); err != nil {
		return false, err
	}
	return true, nil
}

func (r *Redis) Set(ctx context.Context, key string, val any) error {
	buf, err := marshal(val)
	if err != nil {
		return err
	}
	return r.rdb.Set(ctx, key, buf, r.ttl).Err()
}

func (r *Redis) Delete(ctx context.Context, key string) error {
	return r.rdb.Del(ctx, key).Err()
}
