package redis

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/barluscuda/laodvx-server-auth/config"

	"github.com/redis/go-redis/v9"
)

var (
	client *redis.Client
	once   sync.Once
)

func Connect(cfg config.RedisConfig) *redis.Client {
	once.Do(func() {
		client = redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
			Password: cfg.Password,
			DB:       cfg.DB,
		})

		if err := client.Ping(context.Background()).Err(); err != nil {
			log.Fatal("failed to connect to redis:", err)
		}
		log.Println("redis connected")
	})
	return client
}
