package queues

import "github.com/wufe/polo/pkg/models"

type ApplicationInitQueue struct {
	RequestChan  chan *models.Application
	ResponseChan chan error
}

func NewApplicationInit() ApplicationInitQueue {
	return ApplicationInitQueue{
		RequestChan:  make(chan *models.Application),
		ResponseChan: make(chan error),
	}
}

func (q *ApplicationInitQueue) Enqueue(input *models.Application) error {
	q.RequestChan <- input
	return <-q.ResponseChan
}
