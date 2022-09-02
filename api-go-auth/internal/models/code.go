package models

import "time"

type (
	CodeStorage interface {
		CreateAndStore(email, code string, lifetime time.Duration) error
		VerifyCode(email, code string) (bool, error)
	}
)
