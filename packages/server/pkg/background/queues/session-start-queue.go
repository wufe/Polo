package queues

import (
	"github.com/wufe/polo/pkg/models"
)

type SessionStartQueue struct {
	Chan chan SessionStartInput
}

type SessionStartInput struct {
	Session *models.Session
}

func NewSessionStart() SessionStartQueue {
	return SessionStartQueue{
		Chan: make(chan SessionStartInput),
	}
}

func (q *SessionStartQueue) Enqueue(input SessionStartInput) {
	q.Chan <- input
}
