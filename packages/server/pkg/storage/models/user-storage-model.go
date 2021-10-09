package storage_models

import (
	"github.com/wufe/polo/pkg/models"
	"github.com/wufe/polo/pkg/utils"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Name     string
	Email    string
	Password string
	Role     string
}

func (u *User) ToModel() *models.User {
	return &models.User{
		Name:  u.Name,
		Email: u.Email,
		Role:  models.UserRole(u.Role),
	}
}

func GetAdminUser() *User {
	password, _ := utils.GeneratePasswordHash("toor")
	return &User{
		Name:     "Admin",
		Email:    "root@admin.com",
		Password: password,
		Role:     "admin",
	}
}
