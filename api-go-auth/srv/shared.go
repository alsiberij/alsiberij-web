package srv

import (
	"auth/database"
	"auth/logging"
)

var (
	PostgresAuth database.Postgres
	Redis0       database.Redis
	Redis1       database.Redis
	Logger       logging.Logger
)
