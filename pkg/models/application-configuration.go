package models

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/wufe/polo/pkg/utils"
)

type ApplicationConfiguration struct {
	utils.RWLocker        `json:"-"`
	Name                  string            `json:"name"`
	Remote                string            `json:"remote"`
	Target                string            `json:"target"`
	Host                  string            `json:"host"`
	Fetch                 Fetch             `json:"fetch"`
	Watch                 Watch             `json:"watch"`
	IsDefault             bool              `yaml:"is_default" json:"isDefault"`
	Forwards              []Forward         `json:"forwards"`
	Headers               Headers           `json:"headers"`
	Healthcheck           Healthcheck       `json:"healthCheck"`
	Startup               Startup           `json:"startup"`
	Recycle               Recycle           `json:"recycle"`
	Commands              Commands          `json:"commands"`
	MaxConcurrentSessions int               `yaml:"max_concurrent_sessions" json:"maxConcurrentSessions"`
	Port                  PortConfiguration `yaml:"port" json:"port"`
	UseFolderCopy         bool              `yaml:"use_folder_copy" json:"useFolderCopy"`
	CleanOnExit           *bool             `yaml:"clean_on_exit" json:"cleanOnExit" default:"true"`
}

func NewApplicationConfiguration(configuration *ApplicationConfiguration) (*ApplicationConfiguration, error) {
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
		_, err := regexp.Compile(forward.Pattern)
		if err != nil {
			return nil, fmt.Errorf("application.forwards[%d].pattern is not a valid regex: %s", i, err.Error())
		}
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
	return configuration, nil
}

func ConfigurationAreEqual(c1 ApplicationConfiguration, c2 ApplicationConfiguration) bool {
	return reflect.DeepEqual(c1, c2)
}

type Startup struct {
	Timeout int `json:"timeout"`
	Retries int `json:"retries"`
}

type Forward struct {
	Pattern string  `json:"pattern"`
	To      string  `json:"to"`
	Host    string  `json:"host"`
	Headers Headers `json:"headers"`
}

type Watch []string

func (w *Watch) ToSlice() []string {
	ret := []string{}
	for _, e := range *w {
		ret = append(ret, e)
	}
	return ret
}

func (w *Watch) Contains(obj string) bool {
	for _, o := range *w {
		if o == obj {
			return true
		}
	}
	return false
}

type Fetch struct {
	Interval int `json:"interval"`
}

func (a *ApplicationConfiguration) WithLock(f func(*ApplicationConfiguration)) {
	a.Lock()
	defer a.Unlock()
	f(a)
}

func (a *ApplicationConfiguration) WithRLock(f func(*ApplicationConfiguration)) {
	a.RLock()
	defer a.RUnlock()
	f(a)
}
