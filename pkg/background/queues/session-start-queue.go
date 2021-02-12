package queues

import "github.com/wufe/polo/pkg/models"

type SessionStartQueue struct {
	Chan chan *models.Session
}

func NewSessionStart() SessionStartQueue {
	return SessionStartQueue{
		Chan: make(chan *models.Session),
	}
}

func (q *SessionStartQueue) Enqueue(input *models.Session) {
	q.Chan <- input
}
