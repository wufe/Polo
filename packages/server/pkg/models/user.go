package models

const (
	UserRoleStandard UserRole = "standard"
	UserRoleAdmin    UserRole = "admin"
)

type UserRole string

var AllowedRoles []UserRole = []UserRole{
	UserRoleAdmin,
	UserRoleStandard,
}

type User struct {
	Name  string
	Email string
	Role  UserRole
}
