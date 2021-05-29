package services

import (
	"github.com/wufe/polo/pkg/models"
	"github.com/wufe/polo/pkg/storage"
)

type AliasingService struct {
	sessionStorage *storage.Session
}

func NewAliasingService(
	sessionStorage *storage.Session,
) *AliasingService {
	return &AliasingService{
		sessionStorage: sessionStorage,
	}
}

func (a *AliasingService) Next() string {
	names := a.sessionStorage.GetAllSessionsNames()
	return models.NewSessionAlias(names)
}
