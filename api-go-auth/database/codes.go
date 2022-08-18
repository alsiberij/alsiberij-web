package database

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v9"
	"time"
)

const (
	VerificationCodeRedisKey = "VERIFICATION_%s"
)

type (
	Codes struct {
		conn *redis.Client
	}
)

func (r *Codes) Create(email, code string, lifetime time.Duration) error {
	if r.conn == nil {
		return ErrRedisNotInitialized
	}

	return r.conn.Set(context.Background(), fmt.Sprintf(VerificationCodeRedisKey, email), code, lifetime).Err()
}

func (r *Codes) Get(email string) (string, bool, error) {
	if r.conn == nil {
		return "", false, ErrRedisNotInitialized
	}

	response := r.conn.Get(context.Background(), fmt.Sprintf(VerificationCodeRedisKey, email))
	result, err := response.Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", false, nil
		}
		return "", false, err
	}

	return result, true, nil
}
