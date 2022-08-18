package database

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v9"
)

type (
	RedisConfig struct {
		Host     string `json:"host"`
		Port     int    `json:"port"`
		Password string `json:"password"`
		Database int    `json:"database"`
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
			DB:       config.Database,
		}),
	}

	err := r.client.Ping(context.Background()).Err()

	return r, err
}

func (r *Redis) Bans() Bans {
	return Bans{conn: r.client}
}

func (r *Redis) Codes() Codes {
	return Codes{conn: r.client}
}
