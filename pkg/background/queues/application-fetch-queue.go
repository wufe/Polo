package queues

import "github.com/wufe/polo/pkg/models"

type ApplicationFetchQueue struct {
	RequestChan  chan ApplicationFetchInput
	ResponseChan chan error
}

type ApplicationFetchInput struct {
	Application  *models.Application
	WatchObjects bool
}

func NewApplicationFetch() ApplicationFetchQueue {
	return ApplicationFetchQueue{
		RequestChan:  make(chan ApplicationFetchInput),
		ResponseChan: make(chan error),
	}
}

func (q *ApplicationFetchQueue) Enqueue(app *models.Application, watchObjects bool) error {
	q.RequestChan <- ApplicationFetchInput{
		Application:  app,
		WatchObjects: watchObjects,
	}
	return <-q.ResponseChan
}
