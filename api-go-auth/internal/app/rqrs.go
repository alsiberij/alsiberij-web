package app

import (
	"auth/internal/models"
	"regexp"
	"time"
)

const (
	VerificationCodeLength   = 8
	VerificationCodeLifetime = 5 * time.Minute
	EmailMaxLength           = 64
	EmailRegexp              = "^[a-z][a-z\\d-_.]{2,}@[a-z][a-z\\d-]+\\.[a-z][a-z\\d]+$"

	LoginRegexp    = "^[a-z][a-z\\d]{4,32}$"
	PasswordRegexp = "^[\\w!@#$%^&*\\-+=]{8,32}$"

	RefreshTokenRevokeTypeCurrent          = "CURRENT"
	RefreshTokenRevokeTypeAll              = "ALL"
	RefreshTokenRevokeTypeAllExceptCurrent = "ALL_EXCEPT_CURRENT"

	RefreshTokenLength     = 1024
	RefreshTokenAlphabet   = `1234567890abcdef`
	RefreshTokenLifePeriod = 24 * time.Hour

	MinBanReasonLength = 3
	MaxBanReasonLength = 256
	MinBanDuration     = 5 * time.Minute
)

type (
	checkEmailRequest struct {
		Email string `json:"email"`
	}

	registerRequest struct {
		Email    string `json:"email"`
		Code     string `json:"code"`
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	loginRequest struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	refreshRequest struct {
		RefreshToken string `json:"refreshToken"`
	}

	banRequest struct {
		Reason string `json:"reason"`
		Until  int64  `json:"until"`
	}

	changeRoleRequest struct {
		Role string `json:"role"`
	}

	testResponse struct {
		Status bool
	}

	loginResponse struct {
		RefreshToken string `json:"refreshToken"`
	}

	refreshResponse struct {
		AccessToken string `json:"accessToken"`
		ExpiresAt   int64  `json:"expiresAt"`
		IssuedAt    int64  `json:"issuedAt"`
	}
)

var (
	validLogin    = regexp.MustCompile(LoginRegexp).MatchString
	validPassword = regexp.MustCompile(PasswordRegexp).MatchString
	validEmail    = regexp.MustCompile(EmailRegexp).MatchString
	revokeTypes   = []string{RefreshTokenRevokeTypeAll, RefreshTokenRevokeTypeCurrent, RefreshTokenRevokeTypeAllExceptCurrent}
)

func (r *checkEmailRequest) Validate() (*models.Error, error) {
	if !(validEmail(r.Email) && len(r.Email) >= 10 && len(r.Email) <= EmailMaxLength) {
		return models.InvalidEmailError, nil
	}

	return nil, nil
}

func (r *registerRequest) Validate() (*models.Error, error) {
	if !(validEmail(r.Email) && len(r.Email) >= 10 && len(r.Email) <= EmailMaxLength) {
		return models.InvalidEmailError, nil
	}

	if len(r.Code) != VerificationCodeLength {
		return models.InvalidCodeError, nil
	}

	if !validLogin(r.Login) {
		return models.InvalidLoginError, nil
	}

	if !validPassword(r.Password) {
		return models.InvalidPasswordError, nil
	}

	return nil, nil
}

func (r *loginRequest) Validate() (*models.Error, error) {
	if !validLogin(r.Login) {
		return models.WrongCredentialsError, nil
	}

	if !validPassword(r.Password) {
		return models.WrongCredentialsError, nil
	}

	return nil, nil
}

func (r *refreshRequest) Validate() (*models.Error, error) {
	if len(r.RefreshToken) != RefreshTokenLength {
		return models.WrongRefreshTokenError, nil
	}

	return nil, nil
}

func (r *banRequest) Validate() (*models.Error, error) {
	runes := []rune(r.Reason)
	if !(len(runes) >= MinBanReasonLength && len(runes) <= MaxBanReasonLength) {
		return models.InvalidBanReasonError, nil
	}

	t, tn := time.Unix(r.Until, 0), time.Now()
	if !(tn.Before(t) && t.Sub(tn) >= MinBanDuration) {
		return models.InvalidBanTimeError, nil
	}

	return nil, nil
}

func (r *changeRoleRequest) Validate() (*models.Error, error) {
	_, ok := models.ToRole(r.Role)
	if !ok {
		return models.InvalidRoleError, nil
	}
	return nil, nil
}
