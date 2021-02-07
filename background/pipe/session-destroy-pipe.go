package pipe

import "github.com/wufe/polo/models"

type SessionDestroyPipe struct {
	Chan chan *models.Session
}

func NewSessionDestroy() SessionDestroyPipe {
	return SessionDestroyPipe{
		Chan: make(chan *models.Session),
	}
}

func (p *SessionDestroyPipe) Request(session *models.Session) {
	p.Chan <- session
}
