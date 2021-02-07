package request

import (
	"errors"
	"fmt"

	"github.com/wufe/polo/pkg/background"
	"github.com/wufe/polo/pkg/background/pipe"
	"github.com/wufe/polo/pkg/storage"
)

type Service struct {
	isDev              bool
	sessionStorage     *storage.Session
	applicationStorage *storage.Application
	mediator           *background.Mediator
}

var (
	ErrApplicationNotFound error = errors.New("Application not found")
	ErrSessionNotFound     error = errors.New("Session not found")
	ErrSessionIsNotAlive   error = errors.New("Session is not alive")
)

func NewRequestService(
	isDev bool,
	sessionStorage *storage.Session,
	applicationStorage *storage.Application,
	mediator *background.Mediator) *Service {
	return &Service{
		isDev:              isDev,
		sessionStorage:     sessionStorage,
		applicationStorage: applicationStorage,
		mediator:           mediator,
	}
}

func (s *Service) NewSession(checkout string, app string) (*pipe.SessionBuildResult, error) {
	a := s.applicationStorage.Get(app)
	if a == nil {
		return nil, ErrApplicationNotFound
	}
	response := s.mediator.BuildSession.Request(&pipe.SessionBuildInput{
		Checkout:    checkout,
		Application: a,
	})
	if response.Result == pipe.SessionBuildResultFailed {
		return nil, fmt.Errorf("Error requesting new session: %s", response.FailingReason)
	}
	return response, nil
}

func (s *Service) SessionDeletion(uuid string) error {
	session := s.sessionStorage.GetByUUID(uuid)
	if session == nil {
		return ErrSessionNotFound
	}
	if !session.Status.IsAlive() {
		return ErrSessionIsNotAlive
	}
	s.mediator.DestroySession.Request(session)
	return nil
}
