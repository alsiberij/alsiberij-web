package srv

import (
	"auth/database"
	"auth/logging"
)

var (
	PostgresAuth database.Postgres
	Logger       logging.Logger
)
