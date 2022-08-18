package models

import "time"

type (
	User struct {
		//REFERENCE STRUCTURE

		Id        int64
		Email     string
		Role      string
		Login     string
		Password  string
		CreatedAt time.Time
	}
)
