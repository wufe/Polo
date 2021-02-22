package models

import (
	"fmt"
	"os/exec"
	"regexp"
	"time"

	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/wufe/polo/pkg/utils"
)

var (
	ApplicationStatusLoading ApplicationStatus = "loading"
	ApplicationStatusReady   ApplicationStatus = "ready"
)

type Application struct {
	utils.RWLocker          `json:"-"`
	Filename                string `json:"filename"`
	configuration           ApplicationConfiguration
	Status                  ApplicationStatus         `json:"status"`
	Folder                  string                    `json:"folder"`
	BaseFolder              string                    `json:"baseFolder"`
	ObjectsToHashMap        map[string]string         `json:"-"`
	HashToObjectsMap        map[string]*RemoteObject  `json:"-"`
	Branches                map[string]*Branch        `json:"branches"`
	Tags                    []string                  `json:"-"`
	Commits                 []string                  `json:"-"`
	CommitMap               map[string]*object.Commit `json:"-"`
	CompiledForwardPatterns []CompiledForwardPattern  `json:"-"`
}

type ApplicationStatus string

type CompiledForwardPattern struct {
	Pattern *regexp.Regexp
	Forward *Forward
}

type ApplicationCommand struct {
	exec.Cmd
	// TODO: Add "suppressOutputPrint"
}

type ApplicationCommandOutput struct {
	Output   []string
	ExitCode int
}

type RemoteObject struct {
	Branches []string
	Tags     []string
}

type Branch struct {
	Name    string    `json:"name"`
	Hash    string    `json:"hash"`
	Author  string    `json:"author"`
	Date    time.Time `json:"date"`
	Message string    `json:"message"`
}

func NewApplication(configuration *ApplicationConfiguration, filename string) (*Application, error) {
	application := &Application{
		Filename: filename,
		RWLocker: utils.GetMutex(),
		Status:   ApplicationStatusLoading,
	}
	configuration, err := NewApplicationConfiguration(configuration)
	if err != nil {
		return nil, err
	}
	for i, forward := range configuration.Forwards {
		compiledPattern, err := regexp.Compile(forward.Pattern)
		if err != nil {
			return nil, fmt.Errorf("application.forwards[%d].pattern is not a valid regex: %s", i, err.Error())
		}
		application.CompiledForwardPatterns = append(
			application.CompiledForwardPatterns,
			CompiledForwardPattern{
				compiledPattern,
				&forward,
			},
		)
	}
	application.ObjectsToHashMap = make(map[string]string)
	application.HashToObjectsMap = make(map[string]*RemoteObject)
	application.Branches = make(map[string]*Branch)
	application.Tags = []string{}
	application.Commits = []string{}
	application.CommitMap = make(map[string]*object.Commit)
	application.SetConfiguration(*configuration)
	return application, nil
}

func (a *Application) WithLock(f func(*Application)) {
	a.Lock()
	defer a.Unlock()
	f(a)
}

func (a *Application) WithRLock(f func(*Application)) {
	a.RLock()
	defer a.RUnlock()
	f(a)
}

func (a *Application) SetFolder(folder string) {
	a.Lock()
	defer a.Unlock()
	a.Folder = folder
}

func (a *Application) SetBaseFolder(baseFolder string) {
	a.Lock()
	defer a.Unlock()
	a.BaseFolder = baseFolder
}

func (a *Application) GetStatus() ApplicationStatus {
	a.RLock()
	defer a.RUnlock()
	return a.Status
}

func (a *Application) SetStatus(status ApplicationStatus) {
	a.Lock()
	defer a.Unlock()
	a.Status = status
}

func (a *Application) GetConfiguration() ApplicationConfiguration {
	a.RLock()
	defer a.RUnlock()
	return a.configuration
}

func (a *Application) SetConfiguration(conf ApplicationConfiguration) {
	a.Lock()
	defer a.Unlock()
	a.configuration = conf
}
