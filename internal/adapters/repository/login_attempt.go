package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/barluscuda/laodvx-server-auth/config"
	"github.com/barluscuda/laodvx-server-auth/internal/ports"
	"github.com/redis/go-redis/v9"
)

type loginAttemptRepository struct {
	rdb         *redis.Client
	maxAttempts int64
	lockoutTTL  time.Duration
}

func NewLoginAttemptRepository(rdb *redis.Client, cfg config.RateLimitConfig) ports.LoginAttemptRepository {
	return &loginAttemptRepository{
		rdb:         rdb,
		maxAttempts: int64(cfg.LoginMaxAttempts),
		lockoutTTL:  cfg.LoginLockoutTTL,
	}
}

func (r *loginAttemptRepository) IsLocked(key string) (bool, time.Duration, error) {
	ctx := context.Background()
	rkey := loginAttemptKey(key)

	count, err := r.rdb.Get(ctx, rkey).Int64()
	if err == redis.Nil {
		return false, 0, nil
	}
	if err != nil {
		return false, 0, err
	}
	if count < r.maxAttempts {
		return false, 0, nil
	}

	// Account is locked — fetch the remaining TTL from Redis.
	ttl, err := r.rdb.TTL(ctx, rkey).Result()
	if err != nil || ttl <= 0 {
		ttl = r.lockoutTTL
	}
	return true, ttl, nil
}

func (r *loginAttemptRepository) Record(key string) error {
	ctx := context.Background()
	rkey := loginAttemptKey(key)

	count, err := r.rdb.Incr(ctx, rkey).Result()
	if err != nil {
		return err
	}
	// Set the TTL only on the first increment so the window starts then.
	if count == 1 {
		if err := r.rdb.Expire(ctx, rkey, r.lockoutTTL).Err(); err != nil {
			return err
		}
	}
	return nil
}

func (r *loginAttemptRepository) Reset(key string) error {
	return r.rdb.Del(context.Background(), loginAttemptKey(key)).Err()
}

func loginAttemptKey(key string) string {
	return fmt.Sprintf("login:attempts:%s", key)
}
