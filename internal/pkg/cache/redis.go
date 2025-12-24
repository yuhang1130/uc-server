package cache

import (
	"context"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/yuhang1130/gin-server/config"
)

type Redis struct {
	Client *redis.Client
}

func NewRedis(cfg *config.Config) (*Redis, error) {
	opts := &redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	}

	client := redis.NewClient(opts)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}
	log.Println("Connected to Redis successfully.")

	return &Redis{Client: client}, nil
}

func (r *Redis) Close() error {
	return r.Client.Close()
}
