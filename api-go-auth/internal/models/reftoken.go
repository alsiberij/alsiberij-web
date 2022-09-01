package models

import (
	"time"
)

type (
	RefreshToken struct {
		Id         int64
		User       User
		Token      string
		IssuedAt   time.Time
		LastUsedAt time.Time
		IsRevoked  bool
	}

	RefreshTokenStorage interface {
		CreateAndStore(userId int64, tokenValue string) error
		Get(tokenValue string, lifePeriod time.Duration) (*RefreshToken, error)
		Revoke(tokenValue string) error
		RevokeAll(tokenValue string) error
		RevokeAllExceptCurrent(tokenValue string) error
		RevokeAllByUserId(userId int64) error
	}
)
