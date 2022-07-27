package srv

//TODO VALIDATORS
type (
	LoginRequest struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}
	LoginResponse struct {
		RefreshToken string `json:"refreshToken"`
		ExpiresIn    int64  `json:"expiresIn"`
	}

	RefreshRequest struct {
		RefreshToken string `json:"refreshToken"`
	}
	RefreshResponse struct {
		JWT       string `json:"JWT"`
		ExpiresAt int64  `json:"expiresAt"`
		IssuedAt  int64  `json:"issuedAt"`
	}

	CheckEmailRequest struct {
		Email string `json:"email"`
	}

	RegisterRequest struct {
		Email    string `json:"email"`
		Code     int    `json:"code"`
		Login    string `json:"login"`
		Password string `json:"password"`
	}
)
