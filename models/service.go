package models

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/kennygrant/sanitize"
	log "github.com/sirupsen/logrus"
	"github.com/wufe/polo/utils"
)

type Service struct {
	Name                  string            `json:"name"`
	Remote                string            `json:"remote"`
	Target                string            `json:"target"`
	Host                  string            `json:"host"`
	Headers               Headers           `json:"headers"`
	Healthcheck           Healthcheck       `json:"healthCheck"`
	Recycle               Recycle           `json:"recycle"`
	Commands              Commands          `json:"commands"`
	MaxConcurrentSessions int               `yaml:"max_concurrent_sessions" json:"maxConcurrentSessions"`
	Port                  PortConfiguration `yaml:"port" json:"port"`
	ServiceFolder         string            `yaml:"-" json:"serviceFolder"`
	ServiceBaseFolder     string            `yaml:"-" json:"serviceBaseFolder"`
	commandChan           chan *ServiceCommand
	commandResponseChan   chan *ServiceCommandOutput
	rootConfiguration     *RootConfiguration
	objectsToHashMap      map[string]string
	hashToObjectsMap      map[string]*RemoteObject
	branches              []string
	tags                  []string
	commits               []string
	commitMap             map[string]*object.Commit
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
	branches []string
	tags     []string
}

func NewService(service *Service, configuration *RootConfiguration) (*Service, error) {
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
	service.rootConfiguration = configuration
	service.commandChan = make(chan *ServiceCommand)
	service.commandResponseChan = make(chan *ServiceCommandOutput)
	service.objectsToHashMap = make(map[string]string)
	service.hashToObjectsMap = make(map[string]*RemoteObject)
	service.branches = []string{}
	service.tags = []string{}
	service.commits = []string{}
	service.commitMap = make(map[string]*object.Commit)
	return service, nil
}

func (service *Service) Initialize(configuration *RootConfiguration) error {
	sessionsFolder, err := filepath.Abs(configuration.Global.SessionsFolder)
	if err != nil {
		return err
	}
	if _, err := os.Stat(sessionsFolder); os.IsNotExist(err) { // Session folder does not exist
		err := os.Mkdir(sessionsFolder, 0755)
		if err != nil {
			return err
		}
	}
	serviceName := sanitize.Name(service.Name)
	serviceFolder := filepath.Join(sessionsFolder, serviceName)
	if _, err := os.Stat(serviceFolder); os.IsNotExist(err) { // Service folder does not exist
		err := os.Mkdir(serviceFolder, 0755)
		if err != nil {
			return err
		}
	}
	service.ServiceFolder = serviceFolder

	serviceBaseFolder := filepath.Join(serviceFolder, "_base")    // Folder used for performing periodic git fetch --all and/or git log
	if _, err := os.Stat(serviceBaseFolder); os.IsNotExist(err) { // Service folder does not exist
		cmd := exec.Command("git", "clone", service.Remote, "_base")
		cmd.Dir = serviceFolder

		err := utils.ThroughCallback(utils.ExecuteCommand(cmd))(func(line string) {
			log.Infof("[SERVICE:%s (stdout)> ] %s", service.Name, line)
		})

		if err != nil {
			return err
		}

	}
	service.ServiceBaseFolder = serviceBaseFolder

	service.startCommandWatch()
	service.startFetchRoutine()

	return nil
}

func (service *Service) startCommandWatch() {
	go func() {
		for {
			cmd := <-service.commandChan

			output := []string{}

			err := utils.ThroughCallback(utils.ExecuteCommand(&cmd.Cmd))(func(line string) {
				output = append(output, line)
				log.Infof("[SERVICE:%s (stdout)> ] %s", service.Name, line)
			})

			if err != nil {
				log.Errorf("[SERVICE:%s] %s", service.Name, err.Error())
				return
			}

			service.commandResponseChan <- &ServiceCommandOutput{
				Output:   output,
				ExitCode: cmd.ProcessState.ExitCode(),
			}
		}
	}()
}

func (service *Service) defaultErrorLog(err error, except ...error) {
	if err != nil {
		var foundError error
		for _, exceptErr := range except {
			if exceptErr == err {
				foundError = exceptErr
			}
		}
		if foundError == nil {
			log.Errorf("[SERVICE:%s] %s", service.Name, err.Error())
		}
	}
}

func appendWithoutDup(slice []string, elem ...string) {
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
}

func (service *Service) fetchRemote() {

	// Open repository
	repo, err := git.PlainOpen(service.ServiceBaseFolder)
	service.defaultErrorLog(err)
	if err != nil {
		return
	}

	// Fetch
	err = repo.Fetch(&git.FetchOptions{
		RefSpecs: []config.RefSpec{"refs/*:refs/*", "HEAD:refs/heads/HEAD"},
	})
	service.defaultErrorLog(err, git.NoErrAlreadyUpToDate)
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return
	}

	// Get remote
	remote, err := repo.Remote("origin")
	service.defaultErrorLog(err)
	if err != nil {
		return
	}

	// Branches
	refs, err := remote.List(&git.ListOptions{})
	service.defaultErrorLog(err)

	refPrefix := "refs/heads/"
	for _, ref := range refs {
		refName := ref.Name().String()
		if !strings.HasPrefix(refName, refPrefix) {
			continue
		}
		branchName := refName[len(refPrefix):]

		appendWithoutDup(service.branches, branchName)

		service.objectsToHashMap[branchName] = ref.Hash().String()
		service.objectsToHashMap[fmt.Sprintf("origin/%s", branchName)] = ref.Hash().String()
		service.objectsToHashMap[ref.Name().String()] = ref.Hash().String()

		if service.hashToObjectsMap[ref.Hash().String()] == nil {
			service.hashToObjectsMap[ref.Hash().String()] = &RemoteObject{
				branches: []string{},
				tags:     []string{},
			}
		}

		appendWithoutDup(service.hashToObjectsMap[ref.Hash().String()].branches, branchName)
	}

	// Tags
	tags, err := repo.Tags()
	service.defaultErrorLog(err)

	tagPrefix := "refs/tags/"
	err = tags.ForEach(func(ref *plumbing.Reference) error {
		refName := ref.Name().String()
		if !strings.HasPrefix(refName, tagPrefix) {
			return nil
		}
		tagName := refName[len(tagPrefix):]
		service.objectsToHashMap[tagName] = ref.Hash().String()

		appendWithoutDup(service.tags, tagName)
		service.objectsToHashMap[refName] = ref.Hash().String()

		if service.hashToObjectsMap[ref.Hash().String()] == nil {
			service.hashToObjectsMap[ref.Hash().String()] = &RemoteObject{
				branches: []string{},
				tags:     []string{},
			}
		}

		appendWithoutDup(service.hashToObjectsMap[ref.Hash().String()].tags, tagName)

		return nil
	})

	// Log
	// TODO: Configure "since"
	since := time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC)
	until := time.Now().UTC()
	logs, err := repo.Log(&git.LogOptions{All: true, Since: &since, Until: &until, Order: git.LogOrderCommitterTime})
	service.defaultErrorLog(err)
	if err != nil {
		return
	}

	service.commits = []string{}

	err = logs.ForEach(func(commit *object.Commit) error {
		service.objectsToHashMap[commit.Hash.String()] = commit.Hash.String()
		service.commits = append(service.commits, commit.Hash.String())
		service.commitMap[commit.Hash.String()] = commit
		return nil
	})

	log.Infof("[SERVICE:%s] Found %d commits", service.Name, len(service.commits))
}

func (service *Service) startFetchRoutine() {
	go func() {
		for {

			service.fetchRemote()

			time.Sleep(1 * time.Minute)
		}
	}()
}

func (service *Service) ExecCommand(command *ServiceCommand) *ServiceCommandOutput {
	service.commandChan <- command
	return <-service.commandResponseChan
}

func (service *Service) ExecCommandInServiceFolder(command *ServiceCommand) *ServiceCommandOutput {
	command.Dir = service.ServiceFolder
	return service.ExecCommand(command)
}

func (service *Service) ExecCommandInServiceBaseFolder(command *ServiceCommand) *ServiceCommandOutput {
	command.Dir = service.ServiceBaseFolder
	return service.ExecCommand(command)
}

func (service *Service) ExecCommandInServiceCheckoutFolder(command *ServiceCommand, checkout string) *ServiceCommandOutput {
	command.Dir = filepath.Join(service.ServiceFolder, checkout)
	return service.ExecCommand(command)
}
