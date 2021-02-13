package background

import (
	"os"
	"path/filepath"
	"time"

	"github.com/kennygrant/sanitize"
	log "github.com/sirupsen/logrus"
	"github.com/wufe/polo/pkg/models"
	"github.com/wufe/polo/pkg/versioning"
)

type ApplicationInitWorker struct {
	global   *models.GlobalConfiguration
	mediator *Mediator
}

func NewApplicationInitWorker(globalConfiguration *models.GlobalConfiguration, mediator *Mediator) *ApplicationInitWorker {
	worker := &ApplicationInitWorker{
		global:   globalConfiguration,
		mediator: mediator,
	}

	worker.startAcceptingInitRequests()

	return worker
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
	var name string
	application.WithLock(func(a *models.Application) {
		name = a.Name
	})

	log.Infof("[APP:%s] Initializing", name)
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

		auth, err := application.GetAuth()
		if err != nil {
			return err
		}

		gitClient := versioning.GetGitClient(application, auth)

		err = gitClient.Clone(applicationFolder, "_base", application.Remote)
		if err != nil {
			return err
		}

	}
	application.SetBaseFolder(baseFolder)
	w.mediator.ApplicationFetch.Enqueue(application, false)
	w.startApplicationFetchRoutine(application)

	application.SetStatus(models.ApplicationStatusReady)

	return nil
}

func (w *ApplicationInitWorker) startApplicationFetchRoutine(application *models.Application) {
	go func() {
		for {
			time.Sleep(time.Duration(application.Fetch.Interval) * time.Second)

			w.mediator.ApplicationFetch.Enqueue(application, true)
		}
	}()
}

func defaultApplicationErrorLog(application *models.Application, err error, except ...error) {
	if err != nil {
		var foundError error
		for _, exceptErr := range except {
			if exceptErr == err {
				foundError = exceptErr
			}
		}
		if foundError == nil {
			log.Errorf("[APP:%s] %s", application.Name, err.Error())
		}
	}
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
