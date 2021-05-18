package background

import (
	"os"
	"path/filepath"
	"time"

	"github.com/kennygrant/sanitize"
	"github.com/wufe/polo/pkg/logging"
	"github.com/wufe/polo/pkg/models"
	"github.com/wufe/polo/pkg/versioning"
)

type ApplicationInitWorker struct {
	global    *models.GlobalConfiguration
	gitClient versioning.GitClient
	mediator  *Mediator
	log       logging.Logger
}

func NewApplicationInitWorker(globalConfiguration *models.GlobalConfiguration, gitClient versioning.GitClient, mediator *Mediator, logger logging.Logger) *ApplicationInitWorker {
	worker := &ApplicationInitWorker{
		global:    globalConfiguration,
		gitClient: gitClient,
		mediator:  mediator,
		log:       logger,
	}
	return worker
}

func (w *ApplicationInitWorker) Start() {
	w.startAcceptingInitRequests()
}

func (w *ApplicationInitWorker) startAcceptingInitRequests() {
	go func() {
		for {
			application := <-w.mediator.ApplicationInit.RequestChan
			err := w.InitApplication(application)
			w.mediator.ApplicationInit.ResponseChan <- err
		}
	}()
}

func (w *ApplicationInitWorker) InitApplication(application *models.Application) error {
	bus := application.GetEventBus()
	bus.PublishEvent(models.ApplicationEventTypeInitializationStarted, application)
	conf := application.GetConfiguration()
	name := conf.Name
	remote := conf.Remote

	w.log.Infof("[APP:%s] Initializing", name)
	sessionsFolder, err := filepath.Abs(w.global.SessionsFolder)
	if err != nil {
		return err
	}
	if _, err := os.Stat(sessionsFolder); os.IsNotExist(err) { // Session folder does not exist
		err := os.Mkdir(sessionsFolder, 0755)
		if err != nil {
			return err
		}
	}
	applicationName := sanitize.Name(name)
	applicationFolder := filepath.Join(sessionsFolder, applicationName)
	if _, err := os.Stat(applicationFolder); os.IsNotExist(err) { // Application folder does not exist
		err := os.Mkdir(applicationFolder, 0755)
		if err != nil {
			return err
		}
	}
	application.SetFolder(applicationFolder)

	baseFolder := filepath.Join(applicationFolder, "_base") // Folder used for performing periodic git fetch --all and/or git log
	if _, err := os.Stat(baseFolder); os.IsNotExist(err) {  // Application folder does not exist

		err = w.gitClient.Clone(applicationFolder, "_base", remote)
		if err != nil {
			return err
		}

	}
	application.SetBaseFolder(baseFolder)

	w.mediator.ApplicationFetch.Enqueue(application, false)
	w.startApplicationFetchRoutine(application)

	application.SetStatus(models.ApplicationStatusReady)

	bus.PublishEvent(models.ApplicationEventTypeInitializationCompleted, application)

	return nil
}

func (w *ApplicationInitWorker) startApplicationFetchRoutine(application *models.Application) {
	go func() {
		for {
			// Obtaining fetchInterval here because the configuration might change
			conf := application.GetConfiguration()
			fetchInterval := conf.Fetch.Interval
			time.Sleep(time.Duration(fetchInterval) * time.Second)

			w.mediator.ApplicationFetch.Enqueue(application, true)
		}
	}()
}

func appendWithoutDup(slice []string, elem ...string) []string {
	for _, currentElem := range elem {
		foundIndex := -1
		for i, sliceElem := range slice {
			if sliceElem == currentElem {
				foundIndex = i
			}
		}
		if foundIndex == -1 {
			slice = append(slice, currentElem)
		}
	}
	return slice
}
