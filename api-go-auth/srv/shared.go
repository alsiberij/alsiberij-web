package srv

import (
	"auth/logging"
	"auth/repository"
)

var (
	PostgresAuth repository.Postgres
	Logger       logging.Logger
)
