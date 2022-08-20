package srv

import (
	"auth/jwt"
	"auth/utils"
	"regexp"
	"strings"
	"time"
)

type (
	LoginRequest struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	RefreshRequest struct {
		RefreshToken string `json:"refreshToken"`
	}

	CheckEmailRequest struct {
		Email string `json:"email"`
	}

	RegisterRequest struct {
		Email    string `json:"email"`
		Code     string `json:"code"`
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	BanRequest struct {
		Reason string `json:"reason"`
		Until  int64  `json:"until"`
	}

	ChangeRoleRequest struct {
		Role string `json:"role"`
	}
)

var (
	ValidLogin    = regexp.MustCompile("^[a-z][a-z\\d]{4,32}$").MatchString
	ValidPassword = regexp.MustCompile("^[\\w!@#$%^&*\\-+=]{8,32}$").MatchString
	ValidRefresh  = regexp.MustCompile(RefreshTokenAlphabetRegexp).MatchString
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

	if len(r.Code) != VerificationCodeLength {
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

func (r *BanRequest) Validate() (bool, UserMessage) {
	if len([]rune(r.Reason)) < 3 || len([]rune(r.Reason)) > 512 {
		return false, InvalidBanReasonMessage
	}

	if time.Unix(r.Until, 0).Before(time.Now().Add(time.Minute * 5)) {
		return false, InvalidBanTimeMessage
	}

	return true, UserMessage{}
}

func (r *ChangeRoleRequest) Validate() (bool, UserMessage) {
	if !utils.ExistsIn(jwt.AllRoles, r.Role) {
		return false, InvalidRoleUserMessage
	}

	return true, UserMessage{}
}
