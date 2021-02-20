package models

import (
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
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
	Configuration           ApplicationConfiguration  `json:"configuration"`
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

func NewApplication(configuration *ApplicationConfiguration) (*Application, error) {
	application := &Application{}
	application.RWLocker = utils.GetMutex()
	application.Status = ApplicationStatusLoading
	configuration.RWLocker = utils.GetMutex()
	if configuration.Name == "" {
		return nil, errors.New("application.name (required) not defined")
	}
	if configuration.CleanOnExit == nil {
		cleanOnExit := true
		configuration.CleanOnExit = &cleanOnExit
	}
	if configuration.Watch == nil {
		configuration.Watch = []string{}
	}
	if configuration.Remote == "" {
		return nil, errors.New("application.remote (required) not defined; put the git repository URL")
	}
	if configuration.Forwards == nil {
		configuration.Forwards = make([]Forward, 0)
	}
	for i, forward := range configuration.Forwards {
		if forward.Pattern == "" {
			return nil, fmt.Errorf("application.forwards[%d].pattern not defined", i)
		}
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
		if forward.To == "" {
			return nil, fmt.Errorf("application.forwards[%d].to not defined", i)
		}
	}
	if configuration.Fetch.Interval <= 0 {
		configuration.Fetch.Interval = 60
	}
	if configuration.Target == "" {
		configuration.Target = "http://127.0.0.1:{{port}}"
	}
	if configuration.Headers.Add == nil {
		configuration.Headers.Add = []Header{}
	}
	if configuration.Headers.Del == nil {
		configuration.Headers.Del = []string{}
	}
	if configuration.Headers.Set == nil {
		configuration.Headers.Set = []Header{}
	}
	if configuration.Healthcheck.URL == "" {
		configuration.Healthcheck.Method = "GET"
	} else {
		configuration.Healthcheck.Method = strings.ToUpper(configuration.Healthcheck.Method)
	}
	if configuration.Healthcheck.URL == "" {
		configuration.Healthcheck.URL = "/"
	}
	if configuration.Healthcheck.Status == 0 {
		configuration.Healthcheck.Status = 200
	}
	if configuration.Healthcheck.MaxRetries <= 0 {
		configuration.Healthcheck.MaxRetries = 5
	}
	if configuration.Healthcheck.RetryInterval == 0 {
		configuration.Healthcheck.RetryInterval = 30
	}
	if configuration.Healthcheck.RetryTimeout <= 0 {
		configuration.Healthcheck.RetryTimeout = 20 // seconds
	}
	if configuration.Startup.Timeout <= 0 {
		configuration.Startup.Timeout = 300 // seconds
	}
	if configuration.Recycle.InactivityTimeout == 0 {
		configuration.Recycle.InactivityTimeout = 3600 // 1 hour
	}
	if configuration.Commands.Start == nil {
		return nil, errors.New("application.commands.start (required) not defined; put commands required for starting the application; commands accept placeholders")
	}
	for _, command := range configuration.Commands.Start {
		if command.Environment == nil {
			command.Environment = []string{}
		}
	}
	if configuration.Commands.Stop == nil {
		return nil, errors.New("application.commands.stop (required) not defined; put commands required for stopping the application; commands accept placeholders")
	}
	for _, command := range configuration.Commands.Stop {
		if command.Environment == nil {
			command.Environment = []string{}
		}
	}
	if configuration.MaxConcurrentSessions == 0 {
		configuration.MaxConcurrentSessions = 5
	}
	if configuration.Port.Except == nil {
		configuration.Port.Except = []int{}
	}
	application.ObjectsToHashMap = make(map[string]string)
	application.HashToObjectsMap = make(map[string]*RemoteObject)
	application.Branches = make(map[string]*Branch)
	application.Tags = []string{}
	application.Commits = []string{}
	application.CommitMap = make(map[string]*object.Commit)
	application.Configuration = *configuration
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
	return a.Configuration
}
