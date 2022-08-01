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
		IssuedAt   time.Time `json:"issuedAt"`
		LastUsedAt time.Time `json:"lastUsed"`
	}
)
