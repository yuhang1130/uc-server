package cache

import (
	"github.com/go-redis/redis/v8"
	"go.uber.org/fx"
)

func provideRedisClient(r *Redis) *redis.Client {
	return r.Client
}

var Module = fx.Options(
	fx.Provide(NewRedis, provideRedisClient, NewAuthCache, NewUserCache),
)
