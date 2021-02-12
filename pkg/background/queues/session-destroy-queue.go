package queues

import "github.com/wufe/polo/pkg/models"

type SessionDestroyQueue struct {
	Chan chan SessionDestroyInput
}

type SessionDestroyInput struct {
	Session  *models.Session
	Callback func(*models.Session)
}

func NewSessionDestroy() SessionDestroyQueue {
	return SessionDestroyQueue{
		Chan: make(chan SessionDestroyInput),
	}
}

func (q *SessionDestroyQueue) Enqueue(session *models.Session, callback func(*models.Session)) {
	q.Chan <- SessionDestroyInput{
		Session:  session,
		Callback: callback,
	}
}
