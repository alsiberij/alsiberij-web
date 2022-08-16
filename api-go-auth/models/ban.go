package models

import "time"

type (
	Ban struct {
		UserId          int64     `json:"userId"`
		IsActive        bool      `json:"isActive"`
		ActiveUntil     time.Time `json:"activeUntil"`
		CreatedAt       time.Time `json:"createdAt"`
		CreatedByUserId int64     `json:"createdByUserId"`
		Reason          string    `json:"reason"`
	}
)
