package services

import (
	"fmt"

	"github.com/wufe/polo/pkg/background"
	"github.com/wufe/polo/pkg/background/queues"
	"github.com/wufe/polo/pkg/storage"
)

type RequestService struct {
	isDev              bool
	sessionStorage     *storage.Session
	applicationStorage *storage.Application
	mediator           *background.Mediator
}

func NewRequestService(
	isDev bool,
	sessionStorage *storage.Session,
	applicationStorage *storage.Application,
	mediator *background.Mediator) *RequestService {
	return &RequestService{
		isDev:              isDev,
		sessionStorage:     sessionStorage,
		applicationStorage: applicationStorage,
		mediator:           mediator,
	}
}

func (s *RequestService) NewSession(checkout string, app string) (*queues.SessionBuildResult, error) {
	a := s.applicationStorage.Get(app)
	if a == nil {
		return nil, ErrApplicationNotFound
	}
	response := s.mediator.BuildSession.Enqueue(&queues.SessionBuildInput{
		Checkout:    checkout,
		Application: a,
	})
	if response.Result == queues.SessionBuildResultFailed {
		return nil, fmt.Errorf("Error requesting new session: %s", response.FailingReason)
	}
	return response, nil
}

func (s *RequestService) SessionDeletion(uuid string) error {
	session := s.sessionStorage.GetByUUID(uuid)
	if session == nil {
		return ErrSessionNotFound
	}
	if !session.Status.IsAlive() {
		return ErrSessionIsNotAlive
	}
	s.mediator.DestroySession.Enqueue(session, nil)
	return nil
}
