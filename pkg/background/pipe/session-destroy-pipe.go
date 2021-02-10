package pipe

import "github.com/wufe/polo/pkg/models"

type SessionDestroyPipe struct {
	Chan chan SessionDestroyInput
}

type SessionDestroyInput struct {
	Session  *models.Session
	Callback func(*models.Session)
}

func NewSessionDestroy() SessionDestroyPipe {
	return SessionDestroyPipe{
		Chan: make(chan SessionDestroyInput),
	}
}

func (p *SessionDestroyPipe) Request(session *models.Session, callback func(*models.Session)) {
	p.Chan <- SessionDestroyInput{
		Session:  session,
		Callback: callback,
	}
}
