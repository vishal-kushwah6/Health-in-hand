package redis

import (
	configs "artGallery/config"
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

func NewRedisClient(cfg configs.Config) *redis.Client {

	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Redis.HOST, cfg.Redis.PORT),
		DB:       0,
		Password: cfg.Redis.PASSWORD,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		cfg.Logger.Error(context.Background(), "Redis connection failed", map[string]any{
			"error": err.Error(),
		})
	}
	return client
}
