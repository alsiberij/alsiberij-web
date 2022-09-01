package models

type (
	Error struct {
		Message   string
		InnerCode int
	}
)

var (
	AccountIsBannedError = &Error{
		Message:   "Account is banned",
		InnerCode: -1,
	}
	EmailExistsError = &Error{
		Message:   "Email already exists",
		InnerCode: -2,
	}
	LoginExistsError = &Error{
		Message:   "Login already exists",
		InnerCode: -3,
	}
	InvalidEmailError = &Error{
		Message:   "Invalid email",
		InnerCode: -4,
	}
	InvalidCodeError = &Error{
		Message:   "Invalid code",
		InnerCode: -5,
	}
	InvalidLoginError = &Error{
		Message:   "Invalid login",
		InnerCode: -6,
	}
	InvalidPasswordError = &Error{
		Message:   "Invalid password",
		InnerCode: -7,
	}
	InvalidLoginOrPasswordError = &Error{
		Message:   "Invalid login or password",
		InnerCode: -8,
	}
	InvalidRefreshTokenError = &Error{
		Message:   "Invalid refresh token",
		InnerCode: -9,
	}
	InvalidRevokeTypeError = &Error{
		Message:   "Invalid revoke type",
		InnerCode: -10,
	}
	InvalidUserIdError = &Error{
		Message:   "Invalid user id",
		InnerCode: -11,
	}
	InvalidBanReasonError = &Error{
		Message:   "Invalid ban reason",
		InnerCode: -12,
	}
	InvalidBanTimeError = &Error{
		Message:   "Invalid ban time",
		InnerCode: -13,
	}
	InvalidRoleError = &Error{
		Message:   "Invalid role",
		InnerCode: -14,
	}
	InvalidUserIdNoPermission = &Error{
		Message:   "No permission to this user id",
		InnerCode: -15,
	}
	NoAccessError = &Error{
		Message:   "Access denied",
		InnerCode: -16,
	}
)
