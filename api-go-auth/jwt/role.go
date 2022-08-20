package jwt

const (
	RoleCreator        = "CREATOR"
	RoleAdministrator  = "ADMINISTRATOR"
	RoleModerator      = "MODERATOR"
	RolePrivilegedUser = "PRIVILEGED_USER"
	RoleUser           = "USER"
)

var (
	AllRoles = []string{RoleCreator, RoleAdministrator, RoleModerator, RolePrivilegedUser, RoleUser}
)

func RoleIsHigherThan(role1, role2 string) bool {
	switch role1 {
	case RoleCreator:
		return role2 != RoleCreator
	case RoleAdministrator:
		return role2 != RoleCreator && role2 != RoleAdministrator
	case RoleModerator:
		return role2 != RoleCreator && role2 != RoleAdministrator && role2 != RoleModerator
	case RolePrivilegedUser:
		return role2 == RoleUser
	case RoleUser:
		return false
	default:
		return false
	}
}

func RoleIsHigherOrEqualThan(role1, role2 string) bool {
	switch role1 {
	case RoleCreator:
		return true
	case RoleAdministrator:
		return role2 != RoleCreator
	case RoleModerator:
		return role2 != RoleCreator && role2 != RoleAdministrator
	case RolePrivilegedUser:
		return role2 == RolePrivilegedUser || role2 == RoleUser
	case RoleUser:
		return role2 == RoleUser
	default:
		return false
	}
}
