package models

import "time"

type (
	Ban struct {
		BannedUserId    int64
		Reason          string
		ActiveUntil     time.Time
		CreatedByUserId int64
		CreatedAt       time.Time
	}
)
