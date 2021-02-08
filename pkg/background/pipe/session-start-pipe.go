package pipe

import "github.com/wufe/polo/pkg/models"

type SessionStartPipe struct {
	Chan chan *models.Session
}

func NewSessionStart() SessionStartPipe {
	return SessionStartPipe{
		Chan: make(chan *models.Session),
	}
}
