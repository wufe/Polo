package pipe

import "github.com/wufe/polo/pkg/models"

type ApplicationFetchPipe struct {
	RequestChan  chan ApplicationFetchInput
	ResponseChan chan error
}

type ApplicationFetchInput struct {
	Application  *models.Application
	WatchObjects bool
}

func NewApplicationFetch() ApplicationFetchPipe {
	return ApplicationFetchPipe{
		RequestChan:  make(chan ApplicationFetchInput),
		ResponseChan: make(chan error),
	}
}

func (p *ApplicationFetchPipe) Request(app *models.Application, watchObjects bool) error {
	p.RequestChan <- ApplicationFetchInput{
		Application:  app,
		WatchObjects: watchObjects,
	}
	return <-p.ResponseChan
}
