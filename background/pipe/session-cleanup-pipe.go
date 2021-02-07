package pipe

import "github.com/wufe/polo/models"

type SessionCleanupPipe struct {
	Chan chan *SessionCleanupInput
}

func NewSessionCleanup() SessionCleanupPipe {
	return SessionCleanupPipe{
		Chan: make(chan *SessionCleanupInput),
	}
}

func (p *SessionCleanupPipe) Request(input *SessionCleanupInput) {
	p.Chan <- input
}

type SessionCleanupInput struct {
	Session *models.Session
	Status  models.SessionStatus
}
