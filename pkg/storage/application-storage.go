package storage

import (
	"strings"

	"github.com/wufe/polo/pkg/models"
	"github.com/wufe/polo/pkg/utils"
)

// Application is the application storage.
// Contains methods to access and store applications
type Application struct {
	utils.RWLocker
	applications []*models.Application
}

// NewApplication builds new application storage
func NewApplication() *Application {
	return &Application{
		RWLocker:     utils.GetMutex(),
		applications: make([]*models.Application, 0),
	}
}

// Add stores an application
func (a *Application) Add(application *models.Application) {
	a.Lock()
	defer a.Unlock()
	a.applications = append(a.applications, application)
}

// Get retrieves an application by its name.
// If name is an empty string, the "default" application is returned
func (a *Application) Get(name string) *models.Application {
	applications := a.GetAll()
	var foundApplication *models.Application
	for _, application := range applications {
		conf := application.GetConfiguration()
		if name == "" && conf.IsDefault {
			foundApplication = application
			break
		} else if strings.ToLower(conf.Name) == strings.ToLower(name) {
			foundApplication = application
			break
		}
	}
	return foundApplication
}

// GetAll retrieves all applications
func (a *Application) GetAll() []*models.Application {
	a.RLock()
	defer a.RUnlock()
	return a.applications
}
