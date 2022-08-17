package database

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v9"
)

type (
	RedisConfig struct {
		Host     string
		Port     int
		Password string
		Db       int
	}
	Redis struct {
		client *redis.Client
	}
)

func NewRedis(config RedisConfig) (Redis, error) {
	r := Redis{
		client: redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%d", config.Host, config.Port),
			Password: config.Password,
			DB:       config.Db,
		}),
	}

	err := r.client.Ping(context.Background()).Err()

	return r, err
}

func (r *Redis) Bans() Bans {
	return Bans{conn: r.client}
}
