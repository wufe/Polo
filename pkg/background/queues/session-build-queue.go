package queues

import (
	"github.com/wufe/polo/pkg/models"
)

const (
	SessionBuildResultSucceeded    SessionBuildResultType = "succeeded"
	SessionBuildResultAlreadyBuilt SessionBuildResultType = "already_built"
	SessionBuildResultFailed       SessionBuildResultType = "failed"
)

type SessionBuildQueue struct {
	RequestChan  chan *SessionBuildInput
	ResponseChan chan *SessionBuildResult
}

func NewSessionBuild() SessionBuildQueue {
	return SessionBuildQueue{
		RequestChan:  make(chan *SessionBuildInput),
		ResponseChan: make(chan *SessionBuildResult),
	}
}

type SessionBuildInput struct {
	Checkout        string
	Application     *models.Application
	PreviousSession *models.Session
}
type SessionBuildResultType string

type SessionBuildResult struct {
	Result        SessionBuildResultType
	Session       *models.Session
	FailingReason string
	EventBus      *models.SessionLifetimeEventBus
}

func (q *SessionBuildQueue) Enqueue(checkout string, app *models.Application, prevSession *models.Session) *SessionBuildResult {
	q.RequestChan <- &SessionBuildInput{
		Checkout:        checkout,
		Application:     app,
		PreviousSession: prevSession,
	}
	return <-q.ResponseChan
}
