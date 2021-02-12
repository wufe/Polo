package queues

import "github.com/wufe/polo/pkg/models"

type SessionCleanupQueue struct {
	Chan chan *SessionCleanupInput
}

func NewSessionCleanup() SessionCleanupQueue {
	return SessionCleanupQueue{
		Chan: make(chan *SessionCleanupInput),
	}
}

func (q *SessionCleanupQueue) Enqueue(input *SessionCleanupInput) {
	q.Chan <- input
}

type SessionCleanupInput struct {
	Session *models.Session
	Status  models.SessionStatus
}
