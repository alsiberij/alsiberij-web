package database

import (
	"auth/models"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v9"
	"time"
)

const (
	BanRedisKey = "BAN_AUTH_%d"
)

type (
	Bans struct {
		conn *redis.Client
	}
)

func (r *Bans) Create(userId int64, reason string, until int64, byUserId int64) error {
	t := time.Now()
	ban := models.BanDTO{
		UserId:   userId,
		ByUserId: byUserId,
		Reason:   reason,
		At:       t.Unix(),
		Until:    until,
	}
	bytes, _ := json.Marshal(ban)

	return r.conn.Set(context.Background(), fmt.Sprintf(BanRedisKey, userId), bytes, time.Unix(until, 0).Sub(t)).Err()
}

func (r *Bans) Get(userId int64) (models.BanDTO, bool, error) {
	result := r.conn.Get(context.Background(), fmt.Sprintf(BanRedisKey, userId))
	raw, err := result.Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return models.BanDTO{}, false, nil
		}
		return models.BanDTO{}, false, err
	}

	var ban models.BanDTO
	err = json.Unmarshal(raw, &ban)
	return ban, true, err
}

func (r *Bans) Delete(userId int64) error {
	return r.conn.Del(context.Background(), fmt.Sprintf(BanRedisKey, userId)).Err()
}
