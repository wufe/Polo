package pipe

import "github.com/wufe/polo/pkg/models"

type ApplicationInitPipe struct {
	RequestChan  chan *models.Application
	ResponseChan chan error
}

func NewApplicationInit() ApplicationInitPipe {
	return ApplicationInitPipe{
		RequestChan:  make(chan *models.Application),
		ResponseChan: make(chan error),
	}
}

func (p *ApplicationInitPipe) Request(input *models.Application) error {
	p.RequestChan <- input
	return <-p.ResponseChan
}
