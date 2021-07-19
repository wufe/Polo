package queues

import (
	"github.com/wufe/polo/pkg/models"
)

type SessionHealthcheckQueue struct {
	RequestChan  chan SessionHealthcheckInput
	ResponseChan chan struct{}
}

type SessionHealthcheckInput struct {
	Session *models.Session
}

func NewSessionHealthCheck() SessionHealthcheckQueue {
	return SessionHealthcheckQueue{
		RequestChan:  make(chan SessionHealthcheckInput),
		ResponseChan: make(chan struct{}),
	}
}

func (q *SessionHealthcheckQueue) Enqueue(input SessionHealthcheckInput) struct{} {
	q.RequestChan <- input
	return <-q.ResponseChan
}
