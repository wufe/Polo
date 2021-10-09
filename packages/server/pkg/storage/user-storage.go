package storage

import (
	"errors"

	storage_models "github.com/wufe/polo/pkg/storage/models"
	"gorm.io/gorm"
)

type User struct {
	database Database
}

func NewUser(db Database) *User {
	return &User{
		database: db,
	}
}

func (u *User) GetUserByEmail(email string) *storage_models.User {
	var user *storage_models.User
	result := u.database.GetGorm().First(&user, "email = ?", email)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil
	}

	return user
}

func (u *User) AddUser(user *storage_models.User) error {
	result := u.database.GetGorm().Create(user)
	return result.Error
}
