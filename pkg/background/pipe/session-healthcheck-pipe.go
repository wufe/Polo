package pipe

import (
	"github.com/wufe/polo/pkg/models"
)

type SessionHealthcheckPipe struct {
	RequestChan  chan *models.Session
	ResponseChan chan struct{}
}

func NewSessionHealthCheck() SessionHealthcheckPipe {
	return SessionHealthcheckPipe{
		RequestChan:  make(chan *models.Session),
		ResponseChan: make(chan struct{}),
	}
}

func (p *SessionHealthcheckPipe) Request(session *models.Session) struct{} {
	p.RequestChan <- session
	return <-p.ResponseChan
}
