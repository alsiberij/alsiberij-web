package models

const (
	AccountIsBanned               = -1 - iota //Status: 403
	EmailExists                               //Status: 400
	LoginExists                               //Status: 400
	InvalidEmail                              //Status: 400
	InvalidCode                               //Status: 400
	WrongCode                                 //Status: 400
	InvalidLogin                              //Status: 400
	InvalidPassword                           //Status: 400
	WrongCredentials                          //Status: 401
	WrongRefreshToken                         //Status: 401
	InvalidRefreshTokenRevokeType             //Status: 400
	WrongUserId                               //Status: 404
	InvalidBanReason                          //Status: 400
	InvalidBanTime                            //Status: 400
	InvalidMyRole                             //Status: 403
	InvalidRole                               //Status: 400
	NoPermissionToUnbanUser                   //Status: 403
	NoPermissionToBanUser                     //Status: 403
	NoPermissionsToSetThisRole                //Status: 403
	NoPermissionToChangeUserRole              //Status: 403
)

type (
	Error struct {
		Message   string
		InnerCode int
	}
)

var (
	EmailExistsError = &Error{
		Message:   "Email already exists",
		InnerCode: EmailExists,
	}
	LoginExistsError = &Error{
		Message:   "Login already exists",
		InnerCode: LoginExists,
	}
	InvalidEmailError = &Error{
		Message:   "Invalid email",
		InnerCode: InvalidEmail,
	}
	InvalidCodeError = &Error{
		Message:   "Invalid code",
		InnerCode: InvalidCode,
	}
	WrongCodeError = &Error{
		Message:   "Wrong code",
		InnerCode: WrongCode,
	}
	InvalidLoginError = &Error{
		Message:   "Invalid login",
		InnerCode: InvalidLogin,
	}
	InvalidPasswordError = &Error{
		Message:   "Invalid password",
		InnerCode: InvalidPassword,
	}
	WrongCredentialsError = &Error{
		Message:   "Wrong credentials",
		InnerCode: WrongCredentials,
	}
	WrongRefreshTokenError = &Error{
		Message:   "Wrong refresh token",
		InnerCode: WrongRefreshToken,
	}
	InvalidRevokeTypeError = &Error{
		Message:   "Invalid revoke type",
		InnerCode: InvalidRefreshTokenRevokeType,
	}
	WrongUserIdError = &Error{
		Message:   "Wrong user id",
		InnerCode: WrongUserId,
	}
	InvalidBanReasonError = &Error{
		Message:   "Invalid ban reason",
		InnerCode: InvalidBanReason,
	}
	InvalidBanTimeError = &Error{
		Message:   "Invalid ban time",
		InnerCode: InvalidBanTime,
	}
	InvalidRoleError = &Error{
		Message:   "Invalid role",
		InnerCode: InvalidRole,
	}
	InvalidMyRoleError = &Error{
		Message:   "Invalid role",
		InnerCode: InvalidMyRole,
	}
	NoPermissionToBanUserError = &Error{
		Message:   "No permission to ban this user",
		InnerCode: NoPermissionToBanUser,
	}
	NoPermissionToUnbanUserError = &Error{
		Message:   "No permission to unban this user",
		InnerCode: NoPermissionToUnbanUser,
	}
	NoPermissionsToSetThisRoleError = &Error{
		Message:   "No permission to set this role",
		InnerCode: NoPermissionsToSetThisRole,
	}
	NoPermissionToChangeUserRoleError = &Error{
		Message:   "No permission to change user role",
		InnerCode: NoPermissionToChangeUserRole,
	}
)
