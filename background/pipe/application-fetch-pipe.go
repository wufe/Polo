package pipe

import "github.com/wufe/polo/models"

type ApplicationFetchPipe struct {
	RequestChan  chan *models.Application
	ResponseChan chan error
}

func NewApplicationFetch() ApplicationFetchPipe {
	return ApplicationFetchPipe{
		RequestChan:  make(chan *models.Application),
		ResponseChan: make(chan error),
	}
}

func (p *ApplicationFetchPipe) Request(input *models.Application) error {
	p.RequestChan <- input
	return <-p.ResponseChan
}
