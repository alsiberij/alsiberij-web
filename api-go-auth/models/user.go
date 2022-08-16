package models

import "time"

type (
	User struct {
		UserShort
		Password string `json:"password"`
	}

	UserShort struct {
		Id        int64     `json:"id"`
		Email     string    `json:"email"`
		Login     string    `json:"login"`
		Role      string    `json:"role"`
		CreatedAt time.Time `json:"createdAt"`
	}
)
