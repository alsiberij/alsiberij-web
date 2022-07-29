package models

import (
	"time"
)

type (
	RefreshToken struct {
		Id         int64     `json:"id"`
		User       User      `json:"user"`
		Token      string    `json:"token"`
		IsExpired  bool      `json:"isExpired"`
		ExpiresAt  time.Time `json:"expiresAt"`
		IssuedAt   time.Time `json:"issuedAt"`
		LastUsedAt NullTime  `json:"lastUsed"`
	}
)
