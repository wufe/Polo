package models

import (
	"errors"
	"os/exec"
	"strings"
	"time"

	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
)

type Application struct {
	Auth                  Auth                           `json:"-"`
	Name                  string                         `json:"name"`
	Remote                string                         `json:"remote"`
	Target                string                         `json:"target"`
	Host                  string                         `json:"host"`
	IsDefault             bool                           `yaml:"is_default" json:"isDefault"`
	Headers               Headers                        `json:"headers"`
	Healthcheck           Healthcheck                    `json:"healthCheck"`
	Recycle               Recycle                        `json:"recycle"`
	Commands              Commands                       `json:"commands"`
	MaxConcurrentSessions int                            `yaml:"max_concurrent_sessions" json:"maxConcurrentSessions"`
	Port                  PortConfiguration              `yaml:"port" json:"port"`
	UseGitCLI             bool                           `yaml:"use_git_cli" json:"useGitCLI"`
	Folder                string                         `yaml:"-" json:"folder"`
	BaseFolder            string                         `yaml:"-" json:"baseFolder"`
	CommandChan           chan *ApplicationCommand       `yaml:"-" json:"-"`
	CommandResponseChan   chan *ApplicationCommandOutput `yaml:"-" json:"-"`
	ObjectsToHashMap      map[string]string              `yaml:"-" json:"-"`
	HashToObjectsMap      map[string]*RemoteObject       `yaml:"-" json:"-"`
	Branches              map[string]*Branch             `yaml:"-" json:"branches"`
	Tags                  []string                       `yaml:"-" json:"-"`
	Commits               []string                       `yaml:"-" json:"-"`
	CommitMap             map[string]*object.Commit      `yaml:"-" json:"-"`
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

type SSHAuth struct {
	User       string `json:"user"`
	PrivateKey string `yaml:"private_key" json:"privateKey"`
	Password   string `json:"password"`
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
	if application.Remote == "" {
		return nil, errors.New("application.remote (required) not defined; put the git repository URL")
	}
	if application.Target == "" {
		application.Target = "http://127.0.0.1:{{port}}"
	}
	if application.Headers.Add == nil {
		application.Headers.Add = []string{}
	}
	if application.Healthcheck != (Healthcheck{}) {
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
		if application.Healthcheck.RetryInterval == 0 {
			application.Healthcheck.RetryInterval = 30
		}
		if application.Healthcheck.RetryTimeout == 0 {
			application.Healthcheck.RetryTimeout = 300 // 10 minutes
		}
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
	application.CommandChan = make(chan *ApplicationCommand)
	application.CommandResponseChan = make(chan *ApplicationCommandOutput)
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
