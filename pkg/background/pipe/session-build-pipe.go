package pipe

import "github.com/wufe/polo/pkg/models"

const (
	SessionBuildResultSucceeded SessionBuildResultType = "succeeded"
	SessionBuildResultFailed    SessionBuildResultType = "failed"
)

type SessionBuildPipe struct {
	RequestChan  chan *SessionBuildInput
	ResponseChan chan *SessionBuildResult
}

func NewSessionBuild() SessionBuildPipe {
	return SessionBuildPipe{
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

func (p *SessionBuildPipe) Request(input *SessionBuildInput) *SessionBuildResult {
	p.RequestChan <- input
	return <-p.ResponseChan
}
