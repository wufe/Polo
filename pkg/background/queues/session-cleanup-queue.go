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

func (q *SessionCleanupQueue) Enqueue(session *models.Session, status models.SessionStatus) {
	q.Chan <- &SessionCleanupInput{session, status}
}

type SessionCleanupInput struct {
	Session *models.Session
	Status  models.SessionStatus
}
