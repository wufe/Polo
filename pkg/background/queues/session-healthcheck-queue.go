package queues

import (
	"github.com/wufe/polo/pkg/models"
)

type SessionHealthcheckQueue struct {
	RequestChan  chan *models.Session
	ResponseChan chan struct{}
}

func NewSessionHealthCheck() SessionHealthcheckQueue {
	return SessionHealthcheckQueue{
		RequestChan:  make(chan *models.Session),
		ResponseChan: make(chan struct{}),
	}
}

func (q *SessionHealthcheckQueue) Enqueue(session *models.Session) struct{} {
	q.RequestChan <- session
	return <-q.ResponseChan
}
