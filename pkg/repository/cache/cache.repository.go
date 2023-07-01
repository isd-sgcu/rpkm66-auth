package cache

import (
	"github.com/go-redis/redis/v8"
	"github.com/isd-sgcu/rpkm66-auth/internal/repository/cache"
)

type Repository interface {
	SaveCache(key string, value interface{}, ttl int) error
	GetCache(key string, value interface{}) error
}

func NewRepository(client *redis.Client) Repository {
	return cache.NewRepository(client)
}
