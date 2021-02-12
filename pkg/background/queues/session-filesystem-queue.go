package queues

import "github.com/wufe/polo/pkg/models"

type SessionFilesystemQueue struct {
	RequestChan  chan *models.Session
	ResponseChan chan *SessionFilesystemResult
}

func NewSessionFilesystem() SessionFilesystemQueue {
	return SessionFilesystemQueue{
		RequestChan:  make(chan *models.Session),
		ResponseChan: make(chan *SessionFilesystemResult),
	}
}

type SessionFilesystemResult struct {
	CommitFolder string
	Err          error
}

func (q *SessionFilesystemQueue) Enqueue(session *models.Session) *SessionFilesystemResult {
	q.RequestChan <- session
	return <-q.ResponseChan
}
