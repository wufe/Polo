package storage

import (
	"strings"

	"github.com/wufe/polo/models"
)

type Application struct {
	applications []*models.Application
}

func NewApplication() *Application {
	return &Application{
		applications: make([]*models.Application, 0),
	}
}

func (a *Application) Add(application *models.Application) {
	a.applications = append(a.applications, application)
}

func (a *Application) Get(name string) *models.Application {
	var foundApplication *models.Application
	for _, application := range a.applications {
		if name == "" && application.IsDefault {
			foundApplication = application
			break
		} else if strings.ToLower(application.Name) == strings.ToLower(name) {
			foundApplication = application
			break
		}
	}
	return foundApplication
}

func (a *Application) GetAll() []*models.Application {
	return a.applications
}
