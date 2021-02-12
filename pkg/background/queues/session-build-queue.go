package queues

import "github.com/wufe/polo/pkg/models"

const (
	SessionBuildResultSucceeded SessionBuildResultType = "succeeded"
	SessionBuildResultFailed    SessionBuildResultType = "failed"
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
	Checkout    string
	Application *models.Application
}
type SessionBuildResultType string

type SessionBuildResult struct {
	Result        SessionBuildResultType
	Session       *models.Session
	FailingReason string
}

func (q *SessionBuildQueue) Enqueue(input *SessionBuildInput) *SessionBuildResult {
	q.RequestChan <- input
	return <-q.ResponseChan
}
