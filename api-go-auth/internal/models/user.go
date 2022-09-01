package models

import (
	"time"
)

const (
	RoleCreator        UserRole = "CREATOR"
	RoleAdministrator  UserRole = "ADMINISTRATOR"
	RoleModerator      UserRole = "MODERATOR"
	RolePrivilegedUser UserRole = "PRIVILEGED_USER"
	RoleUser           UserRole = "USER"
)

type (
	User struct {
		Id    int64
		Email string
		UserCredentials
		Role      UserRole
		CreatedAt time.Time
	}
	UserCredentials struct {
		Login    string
		Password string
	}

	UserRole string

	UserStorage interface {
		CreateAndStore(email, login, password string) error
		GetByCredentials(credentials UserCredentials) (*User, error)
		GetById(id int64) (*User, error)
		EmailExists(email string) (bool, error)
		LoginExists(login string) (bool, error)
		ChangeRole(id int64, role string) error
	}
)

func (r UserRole) IsHigher(role UserRole) bool {
	switch r {
	case RoleCreator:
		return role != RoleCreator
	case RoleAdministrator:
		return role != RoleCreator && role != RoleAdministrator
	case RoleModerator:
		return role != RoleCreator && role != RoleAdministrator && role != RoleModerator
	case RolePrivilegedUser:
		return role == RoleUser
	case RoleUser:
		return false
	default:
		return false
	}
}

func (r UserRole) IsHigherOrEqual(role UserRole) bool {
	switch r {
	case RoleCreator:
		return true
	case RoleAdministrator:
		return role != RoleCreator
	case RoleModerator:
		return role != RoleCreator && role != RoleAdministrator
	case RolePrivilegedUser:
		return role == RolePrivilegedUser || role == RoleUser
	case RoleUser:
		return role == RoleUser
	default:
		return false
	}
}

func ToRole(role string) (UserRole, bool) {
	userRole := UserRole(role)
	return userRole,
		!(userRole != RoleCreator &&
			userRole != RoleAdministrator &&
			userRole != RoleModerator &&
			userRole != RolePrivilegedUser &&
			userRole != RoleUser)
}
