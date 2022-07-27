package models

import "time"

type (
	User struct {
		Id        int64     `json:"id"`
		Email     string    `json:"email"`
		Login     string    `json:"login"`
		Password  string    `json:"password"`
		Role      string    `json:"role"`
		CreatedAt time.Time `json:"createdAt"`
	}
)
