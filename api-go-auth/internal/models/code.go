package models

import "time"

type (
	CodeStorage interface {
		CreateAndStore(email, code string, lifetime time.Duration) error
		Verify(email, code string) (bool, error)
	}
)
