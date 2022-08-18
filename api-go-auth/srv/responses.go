package srv

import (
	"auth/jwt"
)

type (
	TestResponse struct {
		Status bool `json:"status"`
	}

	LoginResponse struct {
		RefreshToken string `json:"refreshToken"`
	}

	RefreshResponse struct {
		JWT       string `json:"JWT"`
		ExpiresAt int64  `json:"expiresAt"`
		IssuedAt  int64  `json:"issuedAt"`
	}

	ValidateJwtResponse struct {
		JwtClaims jwt.Claims `json:"jwtClaims"`
	}
)
