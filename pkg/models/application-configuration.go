package models

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/jxskiss/base62"
	"github.com/kennygrant/sanitize"
	"github.com/wufe/polo/pkg/models/output"
	"github.com/wufe/polo/pkg/utils"
)

// ApplicationConfiguration contains the configuration of the application
// Usually its retrieval methods override the SharedConfiguration struct
// with the checkout-specific configuration
type ApplicationConfiguration struct {
	SharedConfiguration   `yaml:",inline"` // Base configuration, common for branches and root application configuration
	utils.RWLocker        `json:"-"`
	ID                    string   `json:"id"`
	Name                  string   `json:"name"`
	Hash                  string   `json:"hash"`
	Fetch                 Fetch    `json:"fetch"`
	IsDefault             bool     `yaml:"is_default" json:"isDefault"`
	MaxConcurrentSessions int      `yaml:"max_concurrent_sessions" json:"maxConcurrentSessions"`
	Branches              Branches `yaml:"branches"`
	UseFolderCopy         bool     `yaml:"use_folder_copy" json:"useFolderCopy"`
	CleanOnExit           *bool    `yaml:"clean_on_exit" json:"cleanOnExit" default:"true"`
}

func NewApplicationConfiguration(configuration *ApplicationConfiguration, mutexBuilder utils.MutexBuilder) (*ApplicationConfiguration, error) {
	configuration.RWLocker = mutexBuilder()
	if configuration.Name == "" {
		return nil, errors.New("application.name (required) not defined")
	}
	configuration.ID = sanitize.Name(configuration.Name)

	// Hash generation from sha1 of sanitized name
	configuration.Hash = newConfigurationHashFromAppID(configuration.ID)

	if configuration.CleanOnExit == nil {
		cleanOnExit := true
		configuration.CleanOnExit = &cleanOnExit
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
		if forward.Headers.Add == nil {
			forward.Headers.Add = []Header{}
		}
		if forward.Headers.Del == nil {
			forward.Headers.Del = []string{}
		}
		if forward.Headers.Set == nil {
			forward.Headers.Set = []Header{}
		}
		if forward.Headers.Replace == nil {
			forward.Headers.Replace = []Header{}
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
	if configuration.Headers.Replace == nil {
		configuration.Headers.Replace = []Header{}
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
	if configuration.Healthcheck.Timeout <= 0 {
		configuration.Healthcheck.Timeout = 20 // seconds
	}
	if configuration.Warmup.RetryInterval == 0 {
		configuration.Warmup.RetryInterval = 5
	} else if configuration.Warmup.RetryInterval == -1 {
		configuration.Warmup.RetryInterval = 0
	}
	if configuration.Warmup.URLs == nil {
		configuration.Warmup.URLs = []Warmup{}
	}
	urls := configuration.Warmup.URLs
	for i, u := range urls {
		if u.Status == 0 {
			urls[i].Status = 200
		}
		if u.Method == "" {
			urls[i].Method = "GET"
		} else {
			urls[i].Method = strings.ToUpper(u.Method)
		}
		if u.Timeout <= 0 {
			urls[i].Timeout = 20
		}
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

func (a *ApplicationConfiguration) OverrideWith(override SharedConfiguration) {
	if override.Host != "" {
		a.Host = override.Host
	}
	if override.Remote != "" {
		a.Remote = override.Remote
	}
	if override.Target != "" {
		a.Target = override.Target
	}
	if override.Helper != (Helper{}) {
		if override.Helper.Position != "" {
			a.Helper.Position = override.Helper.Position
		}
	}
	// TODO: Not working yet. Requires compiledForwardPatterns to be stored by session
	if len(override.Forwards) > 0 {
		a.Forwards = override.Forwards
	}
	if len(override.Headers.Add) > 0 {
		a.Headers.Add = override.Headers.Add
	}
	if len(override.Headers.Del) > 0 {
		a.Headers.Del = override.Headers.Del
	}
	if len(override.Headers.Set) > 0 {
		a.Headers.Set = override.Headers.Set
	}
	if len(override.Headers.Replace) > 0 {
		a.Headers.Replace = override.Headers.Replace
	}
	if override.Healthcheck != (Healthcheck{}) {
		a.Healthcheck = override.Healthcheck
	}
	if override.Startup != (Startup{}) {
		if override.Startup.Retries != 0 {
			a.Startup.Retries = override.Startup.Retries
		}
		if override.Startup.Timeout != 0 {
			a.Startup.Timeout = override.Startup.Timeout
		}
	}
	if override.Recycle.InactivityTimeout != 0 {
		a.Recycle.InactivityTimeout = override.Recycle.InactivityTimeout
	}
	if len(override.Commands.Start) > 0 {
		a.Commands.Start = override.Commands.Start
	}
	if len(override.Commands.Stop) > 0 {
		a.Commands.Stop = override.Commands.Stop
	}
	if len(override.Port.Except) > 0 {
		a.Port.Except = override.Port.Except
	}
}

// ToOutput converts this model into an output model
func (a ApplicationConfiguration) ToOutput() output.ApplicationConfiguration {
	return mapApplicationConfiguration(a)
}

func newConfigurationHashFromAppID(appID string) string {
	h := sha1.New()
	h.Write([]byte(appID))
	bs := h.Sum(nil)
	return base62.EncodeToString(bs)[:6]
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

type Fetch struct {
	Interval int `json:"interval"`
}

type Helper struct {
	Position HelperPosition `json:"position"`
}

type HelperPosition string

func (p *HelperPosition) GetStyle() (x string, y string) {
	switch *p {
	case "right-bottom", "bottom-right":
		return "right", "bottom"
	case "right-top", "top-right":
		return "right", "top"
	case "left-top", "top-left":
		return "left", "top"
	case "left-bottom", "bottom-left":
		return "left", "bottom"
	default:
		return "left", "bottom"
	}
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

type RequestConfiguration struct {
	Method  string `json:"method"`
	URL     string `yaml:"url" json:"url"`
	Status  int    `json:"status"`
	Timeout int    `yaml:"timeout" json:"timeout"`
}
