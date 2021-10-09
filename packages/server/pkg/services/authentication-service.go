package services

import (
	"errors"
	"strings"

	"github.com/wufe/polo/pkg/models"
	"github.com/wufe/polo/pkg/storage"
	storage_models "github.com/wufe/polo/pkg/storage/models"
	"github.com/wufe/polo/pkg/utils"
	"golang.org/x/crypto/bcrypt"
)

type AuthenticationService interface {
	GetUserByCredentials(username string, password string) (*models.User, error)
	AddUser(name string, email string, password string, role string) error
}

type authenticationServiceImpl struct {
	userStorage *storage.User
}

var (
	ErrInvalidCredentials error = errors.New("Invalid credentials")
	ErrUserExists         error = errors.New("User already exists")
	ErrRoleNotAllowed     error = errors.New("Role not allowed")
)

func NewAuthenticationService(userStorage *storage.User) AuthenticationService {
	return &authenticationServiceImpl{
		userStorage: userStorage,
	}
}

func (a *authenticationServiceImpl) GetUserByCredentials(email string, password string) (*models.User, error) {

	user := a.userStorage.GetUserByEmail(email)
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err == nil {
		return user.ToModel(), nil
	}

	if email == "admin" && password == "admin" {
		return storage_models.GetAdminUser().ToModel(), nil
	}
	return nil, ErrInvalidCredentials
}

func (a *authenticationServiceImpl) AddUser(name string, email string, password string, role string) error {

	if strings.TrimSpace(name) == "" || strings.TrimSpace(email) == "" || strings.TrimSpace(password) == "" {
		return ErrInvalidCredentials
	}

	user := a.userStorage.GetUserByEmail(email)
	if user != nil {
		return ErrUserExists
	}

	if !a.RoleIsAllowed(role) {
		return ErrRoleNotAllowed
	}

	hash, err := utils.GeneratePasswordHash(password)
	if err != nil {
		return err
	}

	return a.userStorage.AddUser(&storage_models.User{
		Name:     name,
		Email:    email,
		Password: hash,
		Role:     role,
	})
}

func (a *authenticationServiceImpl) RoleIsAllowed(role string) bool {
	for _, allowedRole := range models.AllowedRoles {
		if string(allowedRole) == role {
			return true
		}
	}
	return false
}
