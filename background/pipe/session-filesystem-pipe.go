package pipe

import "github.com/wufe/polo/models"

type SessionFilesystemPipe struct {
	RequestChan  chan *models.Session
	ResponseChan chan *SessionFilesystemResult
}

func NewSessionFilesystem() SessionFilesystemPipe {
	return SessionFilesystemPipe{
		RequestChan:  make(chan *models.Session),
		ResponseChan: make(chan *SessionFilesystemResult),
	}
}

type SessionFilesystemResult struct {
	CommitFolder string
	Err          error
}

func (p *SessionFilesystemPipe) Request(session *models.Session) *SessionFilesystemResult {
	p.RequestChan <- session
	return <-p.ResponseChan
}
