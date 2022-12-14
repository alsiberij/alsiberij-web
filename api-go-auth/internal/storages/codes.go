package storages

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
		querier *redis.Client
	}
)

func NewCodeStorage(q *redis.Client) models.CodeStorage {
	return &CodeStorage{querier: q}
}

func (r *CodeStorage) CreateAndStore(email, code string, lifetime time.Duration) error {
	if r.querier == nil {
		return rds.ErrNotInitialized
	}

	return r.querier.Set(context.Background(), fmt.Sprintf(VerificationCodeRedisKey, email), code, lifetime).Err()
}

func (r *CodeStorage) VerifyCode(email, code string) (bool, error) {
	if r.querier == nil {
		return false, rds.ErrNotInitialized
	}

	result, err := r.querier.Get(context.Background(), fmt.Sprintf(VerificationCodeRedisKey, email)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil
		}
		return false, err
	}

	return result == code, nil
}
