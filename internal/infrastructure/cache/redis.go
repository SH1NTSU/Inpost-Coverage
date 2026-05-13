package cache

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

type Redis struct {
	rdb *redis.Client
	ttl time.Duration
}

func NewRedis(addr string, db int, ttl time.Duration) (*Redis, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:         addr,
		DB:           db,
		DialTimeout:  2 * time.Second,
		ReadTimeout:  500 * time.Millisecond,
		WriteTimeout: 500 * time.Millisecond,
	})
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
