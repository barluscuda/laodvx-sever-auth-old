package redisstore

import (
	"context"
	"fmt"
	"time"

	"github.com/barluscuda/laodvx-server-auth/config"
	"github.com/barluscuda/laodvx-server-auth/internal/ports"

	"github.com/redis/go-redis/v9"
)

type registerAttemptRepository struct {
	rdb         *redis.Client
	maxAttempts int64
	lockoutTTL  time.Duration
}

func NewRegisterAttemptRepository(rdb *redis.Client, cfg config.EmailConfig) ports.RegistrationAttemptRepository {
	return &registerAttemptRepository{
		rdb:         rdb,
		maxAttempts: int64(cfg.OTPMaxVerifyAttempts),
		lockoutTTL:  cfg.OTPVerifyLockoutTTL,
	}
}

func (r *registerAttemptRepository) IsLocked(key string) (bool, time.Duration, error) {
	ctx := context.Background()
	rkey := registerAttemptKey(key)

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

	ttl, err := r.rdb.TTL(ctx, rkey).Result()
	if err != nil || ttl <= 0 {
		ttl = r.lockoutTTL
	}
	return true, ttl, nil
}

func (r *registerAttemptRepository) Record(key string) error {
	ctx := context.Background()
	rkey := registerAttemptKey(key)

	count, err := r.rdb.Incr(ctx, rkey).Result()
	if err != nil {
		return err
	}
	if count == 1 {
		if err := r.rdb.Expire(ctx, rkey, r.lockoutTTL).Err(); err != nil {
			return err
		}
	}
	return nil
}

func (r *registerAttemptRepository) Reset(key string) error {
	return r.rdb.Del(context.Background(), registerAttemptKey(key)).Err()
}

func registerAttemptKey(key string) string {
	return fmt.Sprintf("register:verify:attempts:%s", key)
}
