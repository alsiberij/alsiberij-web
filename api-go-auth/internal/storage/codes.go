package storage

import (
	"auth/internal/models"
	"auth/pkg/rds"
	"context"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v9"
	"time"
)

//TODO context

const (
	VerificationCodeRedisKey = "VERIFICATION_EMAIL_%s"
)

type (
	CodeStorage struct {
		conn *redis.Client
	}
)

func NewCodeStorage(q *redis.Client) models.CodeStorage {
	return &CodeStorage{conn: q}
}

func (r *CodeStorage) CreateAndStore(email, code string, lifetime time.Duration) error {
	if r.conn == nil {
		return rds.ErrNotInitialized
	}

	return r.conn.Set(context.Background(), fmt.Sprintf(VerificationCodeRedisKey, email), code, lifetime).Err()
}

func (r *CodeStorage) Verify(email, code string) (bool, error) {
	if r.conn == nil {
		return false, rds.ErrNotInitialized
	}

	response := r.conn.Get(context.Background(), fmt.Sprintf(VerificationCodeRedisKey, email))
	result, err := response.Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil
		}
		return false, err
	}

	return result == code, nil
}
