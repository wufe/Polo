package services

import (
	"fmt"

	"github.com/wufe/polo/pkg/background"
	"github.com/wufe/polo/pkg/background/queues"
	"github.com/wufe/polo/pkg/models"
	"github.com/wufe/polo/pkg/storage"
	"github.com/wufe/polo/pkg/utils"
)

type RequestService struct {
	isDev              bool
	sessionStorage     *storage.Session
	applicationStorage *storage.Application
	mediator           *background.Mediator
}

func NewRequestService(
	environment utils.Environment,
	sessionStorage *storage.Session,
	applicationStorage *storage.Application,
	mediator *background.Mediator) *RequestService {
	return &RequestService{
		isDev:              environment.IsDev(),
		sessionStorage:     sessionStorage,
		applicationStorage: applicationStorage,
		mediator:           mediator,
	}
}

// NewSession requests for a new session to be built
// at a specific checkout (can be a commit ID, a branch name or a tag)
// for a specific app
// If detectBranchOrTag is set to true, if the checkout is a commit ID,
// the builder will try to detect if the commit belongs to a branch or a tag
func (s *RequestService) NewSession(checkout string, app string, detectBranchOrTag bool) (*queues.SessionBuildResult, error) {
	a := s.applicationStorage.Get(app)
	if a == nil {
		return nil, ErrApplicationNotFound
	}
	response := s.mediator.BuildSession.Enqueue(checkout, a, nil, nil, detectBranchOrTag)
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
	session.SetKillReason(models.KillReasonStopped)
	s.mediator.DestroySession.Enqueue(session, nil)
	return nil
}
