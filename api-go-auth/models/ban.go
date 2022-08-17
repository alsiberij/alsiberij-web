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
	BanDTO struct {
		UserId   int64  `json:"userId"`
		ByUserId int64  `json:"byUserId"`
		Reason   string `json:"reason"`
		At       int64  `json:"at"`
		Until    int64  `json:"until"`
	}
)
