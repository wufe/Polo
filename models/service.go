package models

import (
	"errors"
	"os/exec"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
)

type Service struct {
	Auth                  Auth                       `json:"-"`
	Name                  string                     `json:"name"`
	Remote                string                     `json:"remote"`
	Target                string                     `json:"target"`
	Host                  string                     `json:"host"`
	Headers               Headers                    `json:"headers"`
	Healthcheck           Healthcheck                `json:"healthCheck"`
	Recycle               Recycle                    `json:"recycle"`
	Commands              Commands                   `json:"commands"`
	MaxConcurrentSessions int                        `yaml:"max_concurrent_sessions" json:"maxConcurrentSessions"`
	Port                  PortConfiguration          `yaml:"port" json:"port"`
	ServiceFolder         string                     `yaml:"-" json:"serviceFolder"`
	ServiceBaseFolder     string                     `yaml:"-" json:"serviceBaseFolder"`
	CommandChan           chan *ServiceCommand       `yaml:"-" json:"-"`
	CommandResponseChan   chan *ServiceCommandOutput `yaml:"-" json:"-"`
	ObjectsToHashMap      map[string]string          `yaml:"-" json:"-"`
	HashToObjectsMap      map[string]*RemoteObject   `yaml:"-" json:"-"`
	Branches              []string                   `yaml:"-" json:"-"`
	Tags                  []string                   `yaml:"-" json:"-"`
	Commits               []string                   `yaml:"-" json:"-"`
	CommitMap             map[string]*object.Commit  `yaml:"-" json:"-"`
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

type ServiceCommand struct {
	exec.Cmd
	// TODO: Add "suppressOutputPrint"
}

type ServiceCommandOutput struct {
	Output   []string
	ExitCode int
}

type RemoteObject struct {
	Branches []string
	Tags     []string
}

func NewService(service *Service) (*Service, error) {
	// Service %s configuration error:
	if service.Name == "" {
		return nil, errors.New("service.name (required) not defined")
	}
	if service.Remote == "" {
		return nil, errors.New("service.remote (required) not defined; put the git repository URL")
	}
	if service.Target == "" {
		return nil, errors.New("service.target (required) not defined; put the target application URL; accepts placeholders")
	}
	if service.Headers.Add == nil {
		service.Headers.Add = []string{}
	}
	if service.Healthcheck.URL == "" {
		service.Healthcheck.Method = "GET"
	} else {
		service.Healthcheck.Method = strings.ToUpper(service.Healthcheck.Method)
	}
	if service.Healthcheck.URL == "" {
		service.Healthcheck.URL = "/"
	}
	if service.Healthcheck.Status == 0 {
		service.Healthcheck.Status = 200
	}
	if service.Healthcheck.RetryInterval == 0 {
		service.Healthcheck.RetryInterval = 30
	}
	if service.Healthcheck.RetryTimeout == 0 {
		service.Healthcheck.RetryTimeout = 300 // 10 minutes
	}
	if service.Recycle.InactivityTimeout == 0 {
		service.Recycle.InactivityTimeout = 3600 // 1 hour
	}
	if service.Commands.Start == nil {
		return nil, errors.New("service.commands.start (required) not defined; put commands required for starting the application; commands accept placeholders")
	}
	for _, command := range service.Commands.Start {
		if command.Environment == nil {
			command.Environment = []string{}
		}
	}
	if service.Commands.Stop == nil {
		return nil, errors.New("service.commands.stop (required) not defined; put commands required for stopping the application; commands accept placeholders")
	}
	for _, command := range service.Commands.Stop {
		if command.Environment == nil {
			command.Environment = []string{}
		}
	}
	if service.MaxConcurrentSessions == 0 {
		service.MaxConcurrentSessions = 5
	}
	if service.Port.Except == nil {
		service.Port.Except = []int{}
	}
	if service.Auth.SSH != (SSHAuth{}) {
		if service.Auth.SSH.User == "" {
			service.Auth.SSH.User = "git"
		}
	}
	service.CommandChan = make(chan *ServiceCommand)
	service.CommandResponseChan = make(chan *ServiceCommandOutput)
	service.ObjectsToHashMap = make(map[string]string)
	service.HashToObjectsMap = make(map[string]*RemoteObject)
	service.Branches = []string{}
	service.Tags = []string{}
	service.Commits = []string{}
	service.CommitMap = make(map[string]*object.Commit)
	return service, nil
}

func (service *Service) GetAuth() (transport.AuthMethod, error) {
	if service.Auth == (Auth{}) {
		return nil, nil
	}
	if service.Auth.Basic != (BasicAuth{}) {
		return &http.BasicAuth{
			Username: service.Auth.Basic.Username,
			Password: service.Auth.Basic.Password,
		}, nil
	}
	if service.Auth.Token != "" {
		return &http.BasicAuth{
			Username: "_",
			Password: service.Auth.Token,
		}, nil
	}
	if service.Auth.SSH != (SSHAuth{}) {
		publicKeys, err := ssh.NewPublicKeysFromFile(
			service.Auth.SSH.User,
			service.Auth.SSH.PrivateKey,
			service.Auth.SSH.Password,
		)
		if err != nil {
			return nil, err
		}
		return publicKeys, nil
	}
	return nil, nil
}
