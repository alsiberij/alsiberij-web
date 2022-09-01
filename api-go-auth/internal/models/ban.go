package models

import "time"

type (
	Ban struct {
		UserId   int64
		ByUserId int64
		Reason   string
		At       time.Time
		Until    time.Time
	}

	BanStorage interface {
		CreateAndStore(userId int64, reason string, until int64, byUserId int64) error
		Get(userId int64) (*Ban, error)
		Delete(userId int64) error
	}
)
