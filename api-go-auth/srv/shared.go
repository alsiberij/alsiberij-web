package srv

import (
	"auth/database"
	"auth/logging"
	"auth/utils"
	"time"
)

const (
	RefreshTokenLength         = uint(1024)
	RefreshTokenAlphabet       = `<->`
	RefreshTokenAlphabetRegexp = `^[\<\-\>]+$`
	RefreshTokenExpirationTime = 24 * time.Hour

	RefreshTokenRevokeTypeCurrent          = "CURRENT"
	RefreshTokenRevokeTypeAll              = "ALL"
	RefreshTokenRevokeTypeAllExceptCurrent = "ALL_EXCEPT_CURRENT"

	VerificationCodeLength   = 8
	VerificationCodeLifetime = 5 * time.Minute
)

var (
	PostgresAuth database.Postgres
	Redis0       database.Redis
	Redis1       database.Redis
	Logger       logging.Logger
	Random       = utils.NewRandom(time.Now().Unix())
)
