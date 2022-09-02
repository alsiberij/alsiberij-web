package rds

import (
	"context"
	"errors"
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

var (
	ErrNotInitialized = errors.New("nil db")
)

func NewRedis(config RedisConfig) (*Redis, error) {
	r := Redis{
		client: redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%d", config.Host, config.Port),
			Password: config.Password,
			DB:       config.Database,
		}),
	}

	err := r.client.Ping(context.Background()).Err()
	if err != nil {
		return nil, err
	}

	return &r, err
}

func (r *Redis) Client() *redis.Client {
	return r.client
}

func (r *Redis) Close() {
	if r.client != nil {
		_ = r.Client().Close()
	}
}
