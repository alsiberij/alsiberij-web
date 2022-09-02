package storage

import (
	"auth/internal/models"
	"auth/pkg/rds"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v9"
	"time"
)

//TODO context

const (
	BanRedisKey = "BAN_AUTH_%d"
)

type (
	BanStorage struct {
		conn *redis.Client
	}

	banSerialized struct {
		UserId   int64  `json:"userId"`
		ByUserId int64  `json:"byUserId"`
		Reason   string `json:"reason"`
		At       int64  `json:"at"`
		Until    int64  `json:"until"`
	}
)

func NewBanStorage(q *redis.Client) models.BanStorage {
	return &BanStorage{conn: q}
}

func (r *BanStorage) CreateAndStore(userId int64, reason string, until int64, byUserId int64) error {
	if r.conn == nil {
		return rds.ErrNotInitialized
	}

	t := time.Now()
	ban := banSerialized{
		UserId:   userId,
		ByUserId: byUserId,
		Reason:   reason,
		At:       t.Unix(),
		Until:    until,
	}

	bytes, _ := json.Marshal(ban)

	return r.conn.Set(context.Background(), fmt.Sprintf(BanRedisKey, userId), bytes, time.Unix(until, 0).Sub(t)).Err()
}

func (r *BanStorage) Get(userId int64) (*models.Ban, error) {
	if r.conn == nil {
		return nil, rds.ErrNotInitialized
	}

	result := r.conn.Get(context.Background(), fmt.Sprintf(BanRedisKey, userId))
	raw, err := result.Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}

	var ban banSerialized
	err = json.Unmarshal(raw, &ban)
	if err != nil {
		return nil, err
	}

	return &models.Ban{
		UserId:   ban.UserId,
		ByUserId: ban.ByUserId,
		Reason:   ban.Reason,
		At:       time.Unix(ban.At, 0),
		Until:    time.Unix(ban.Until, 0),
	}, nil
}

func (r *BanStorage) Delete(userId int64) error {
	if r.conn == nil {
		return rds.ErrNotInitialized
	}

	return r.conn.Del(context.Background(), fmt.Sprintf(BanRedisKey, userId)).Err()
}
