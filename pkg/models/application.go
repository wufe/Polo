package models

import (
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
)

type Application struct {
	Auth                    Auth                      `json:"-"`
	Name                    string                    `json:"name"`
	Remote                  string                    `json:"remote"`
	Target                  string                    `json:"target"`
	Host                    string                    `json:"host"`
	Fetch                   Fetch                     `json:"fetch"`
	Watch                   Watch                     `json:"watch"`
	IsDefault               bool                      `yaml:"is_default" json:"isDefault"`
	Forwards                []Forward                 `json:"forwards"`
	Headers                 Headers                   `json:"headers"`
	Healthcheck             Healthcheck               `json:"healthCheck"`
	Startup                 Startup                   `json:"startup"`
	Recycle                 Recycle                   `json:"recycle"`
	Commands                Commands                  `json:"commands"`
	MaxConcurrentSessions   int                       `yaml:"max_concurrent_sessions" json:"maxConcurrentSessions"`
	Port                    PortConfiguration         `yaml:"port" json:"port"`
	UseGitCLI               bool                      `yaml:"use_git_cli" json:"useGitCLI"`
	Folder                  string                    `yaml:"-" json:"folder"`
	BaseFolder              string                    `yaml:"-" json:"baseFolder"`
	ObjectsToHashMap        map[string]string         `yaml:"-" json:"-"`
	HashToObjectsMap        map[string]*RemoteObject  `yaml:"-" json:"-"`
	Branches                map[string]*Branch        `yaml:"-" json:"branches"`
	Tags                    []string                  `yaml:"-" json:"-"`
	Commits                 []string                  `yaml:"-" json:"-"`
	CommitMap               map[string]*object.Commit `yaml:"-" json:"-"`
	CompiledForwardPatterns []CompiledForwardPattern  `yaml:"-" json:"-"`
}

type Auth struct {
	Basic BasicAuth `json:"basic"`
	Token string    `json:"token"`
	SSH   SSHAuth   `yaml:"ssh" json:"ssh"`
}

type BasicAuth struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Startup struct {
	Timeout int `json:"timeout"`
}

type SSHAuth struct {
	User       string `json:"user"`
	PrivateKey string `yaml:"private_key" json:"privateKey"`
	Password   string `json:"password"`
}

type Forward struct {
	Pattern string  `json:"pattern"`
	To      string  `json:"to"`
	Host    string  `json:"host"`
	Headers Headers `json:"headers"`
}

type Watch []string

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

func NewApplication(application *Application) (*Application, error) {
	if application.Name == "" {
		return nil, errors.New("application.name (required) not defined")
	}
	if application.Watch == nil {
		application.Watch = []string{}
	}
	if application.Remote == "" {
		return nil, errors.New("application.remote (required) not defined; put the git repository URL")
	}
	if application.Forwards == nil {
		application.Forwards = make([]Forward, 0)
	}
	for i, forward := range application.Forwards {
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
	if application.Fetch.Interval <= 0 {
		application.Fetch.Interval = 60
	}
	if application.Target == "" {
		application.Target = "http://127.0.0.1:{{port}}"
	}
	if application.Headers.Add == nil {
		application.Headers.Add = []Header{}
	}
	if application.Headers.Del == nil {
		application.Headers.Del = []string{}
	}
	if application.Headers.Set == nil {
		application.Headers.Set = []Header{}
	}
	if application.Healthcheck.URL == "" {
		application.Healthcheck.Method = "GET"
	} else {
		application.Healthcheck.Method = strings.ToUpper(application.Healthcheck.Method)
	}
	if application.Healthcheck.URL == "" {
		application.Healthcheck.URL = "/"
	}
	if application.Healthcheck.Status == 0 {
		application.Healthcheck.Status = 200
	}
	if application.Healthcheck.MaxRetries <= 0 {
		application.Healthcheck.MaxRetries = 10
	}
	if application.Healthcheck.RetryInterval == 0 {
		application.Healthcheck.RetryInterval = 30
	}
	if application.Healthcheck.RetryTimeout <= 0 {
		application.Healthcheck.RetryTimeout = 20 // seconds
	}
	if application.Startup.Timeout <= 0 {
		application.Startup.Timeout = 300 // seconds
	}
	if application.Recycle.InactivityTimeout == 0 {
		application.Recycle.InactivityTimeout = 3600 // 1 hour
	}
	if application.Commands.Start == nil {
		return nil, errors.New("application.commands.start (required) not defined; put commands required for starting the application; commands accept placeholders")
	}
	for _, command := range application.Commands.Start {
		if command.Environment == nil {
			command.Environment = []string{}
		}
	}
	if application.Commands.Stop == nil {
		return nil, errors.New("application.commands.stop (required) not defined; put commands required for stopping the application; commands accept placeholders")
	}
	for _, command := range application.Commands.Stop {
		if command.Environment == nil {
			command.Environment = []string{}
		}
	}
	if application.MaxConcurrentSessions == 0 {
		application.MaxConcurrentSessions = 5
	}
	if application.Port.Except == nil {
		application.Port.Except = []int{}
	}
	if application.Auth.SSH != (SSHAuth{}) {
		if application.Auth.SSH.User == "" {
			application.Auth.SSH.User = "git"
		}
	}
	application.ObjectsToHashMap = make(map[string]string)
	application.HashToObjectsMap = make(map[string]*RemoteObject)
	application.Branches = make(map[string]*Branch)
	application.Tags = []string{}
	application.Commits = []string{}
	application.CommitMap = make(map[string]*object.Commit)
	return application, nil
}

func (application *Application) GetAuth() (transport.AuthMethod, error) {
	if application.Auth == (Auth{}) {
		return nil, nil
	}
	if application.Auth.Basic != (BasicAuth{}) {
		return &http.BasicAuth{
			Username: application.Auth.Basic.Username,
			Password: application.Auth.Basic.Password,
		}, nil
	}
	if application.Auth.Token != "" {
		return &http.BasicAuth{
			Username: "_",
			Password: application.Auth.Token,
		}, nil
	}
	if application.Auth.SSH != (SSHAuth{}) {
		publicKeys, err := ssh.NewPublicKeysFromFile(
			application.Auth.SSH.User,
			application.Auth.SSH.PrivateKey,
			application.Auth.SSH.Password,
		)
		if err != nil {
			return nil, err
		}
		return publicKeys, nil
	}
	return nil, nil
}
