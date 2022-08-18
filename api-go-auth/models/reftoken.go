package models

import (
	"time"
)

type (
	RefreshToken struct {
		//REFERENCE STRUCTURE

		Id         int64
		User       User
		Token      string
		IssuedAt   time.Time
		LastUsedAt time.Time
	}

	RefreshTokenWithUserData struct {
		Id       int64
		UserId   int64
		UserRole string
	}
)
