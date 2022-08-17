package srv

import (
	"auth/database"
	"auth/logging"
)

var (
	PostgresAuth database.Postgres
	Redis        database.Redis
	Logger       logging.Logger
)
