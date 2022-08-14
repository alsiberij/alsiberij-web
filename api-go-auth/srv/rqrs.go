package srv

import (
	"auth/jwt"
	"auth/models"
	"fmt"
	"regexp"
	"strings"
)

type (
	TestResponse struct {
		Status bool `json:"status"`
	}

	LoginRequest struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}
	LoginResponse struct {
		RefreshToken string `json:"refreshToken"`
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

	UsersResponse struct {
		Count int `json:"count"`
		List  []models.UserShort
	}

	ChangeUserStatusRequest struct {
		IsBanned bool `json:"isBanned"`
	}

	ValidateJwtResponse struct {
		JwtClaims jwt.Claims `json:"jwtClaims"`
	}
)

var (
	ValidLogin    = regexp.MustCompile("^[a-z]{6,32}$").MatchString
	ValidPassword = regexp.MustCompile("^[\\w!@#$%^&*\\-+=]{6,32}$").MatchString
	ValidRefresh  = regexp.MustCompile(fmt.Sprintf("^[%s]+$", RefreshTokenAlphabet)).MatchString
	ValidEmail    = regexp.MustCompile("^[a-z][a-z\\d-_.]{2,}@[a-z][a-z\\d-]+\\.[a-z][a-z\\d]+$").MatchString
)

func (r *LoginRequest) Validate() (bool, UserMessage) {
	r.Login = strings.ToLower(r.Login)
	if !ValidLogin(r.Login) {
		return false, InvalidLoginUserMessage
	}

	if !ValidPassword(r.Password) {
		return false, InvalidPasswordUserMessage
	}

	return true, UserMessage{}
}

func (r *RefreshRequest) Validate() (bool, UserMessage) {
	if !ValidRefresh(r.RefreshToken) || len(r.RefreshToken) != 1024 {
		return false, InvalidRefreshTokenUserMessage
	}

	return true, UserMessage{}
}

func (r *CheckEmailRequest) Validate() (bool, UserMessage) {
	r.Email = strings.ToLower(r.Email)
	if !ValidEmail(r.Email) || len(r.Email) > 64 {
		return false, InvalidEmailUserMessage
	}

	return true, UserMessage{}
}

func (r *RegisterRequest) Validate() (bool, UserMessage) {
	r.Email = strings.ToLower(r.Email)
	if !ValidEmail(r.Email) || len(r.Email) > 64 {
		return false, InvalidEmailUserMessage
	}

	if r.Code < 100_000 || r.Code > 999_999 {
		return false, InvalidCodeUserMessage
	}

	r.Login = strings.ToLower(r.Login)
	if !ValidLogin(r.Login) {
		return false, InvalidLoginUserMessage
	}

	if !ValidPassword(r.Password) {
		return false, InvalidPasswordUserMessage
	}

	return true, UserMessage{}
}
