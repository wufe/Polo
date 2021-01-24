package models

import (
	"bufio"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/kennygrant/sanitize"
	log "github.com/sirupsen/logrus"
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
}

type ServiceCommand struct {
	exec.Cmd
	// TODO: Add "suppressOutputPrint"
}

type ServiceCommandOutput struct {
	Output   string
	ExitCode int
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
		service.Healthcheck.RetryTimeout = 600 // 10 minutes
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
	service.commandChan = make(chan *ServiceCommand)
	service.commandResponseChan = make(chan *ServiceCommandOutput)
	service.rootConfiguration = configuration
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
		errPipe, err := cmd.StderrPipe()
		if err != nil {
			return err
		}
		outPipe, err := cmd.StdoutPipe()
		if err != nil {
			return err
		}
		if err := cmd.Start(); err != nil {
			return err
		}

		go func() {
			scanner := bufio.NewScanner(errPipe)
			scanner.Split(bufio.ScanLines)
			for scanner.Scan() {
				line := scanner.Text()
				log.Infof("[SERVICE:%s (stderr)> ] %s", service.Name, line)
			}
		}()
		go func() {
			scanner := bufio.NewScanner(outPipe)
			scanner.Split(bufio.ScanLines)
			for scanner.Scan() {
				line := scanner.Text()
				log.Infof("[SERVICE:%s (stdout)> ] %s", service.Name, line)
			}
		}()

		cmd.Wait()
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
			errPipe, err := cmd.StderrPipe()
			if err != nil {
				log.Errorf("[SERVICE:%s] %s", service.Name, err.Error())
				return
			}
			outPipe, err := cmd.StdoutPipe()
			if err != nil {
				log.Errorf("[SERVICE:%s] %s", service.Name, err.Error())
				return
			}
			if err := cmd.Start(); err != nil {
				log.Errorf("[SERVICE:%s] %s", service.Name, err.Error())
				return
			}

			// TODO: Check race condition between these two
			output := ""
			go func() {
				scanner := bufio.NewScanner(errPipe)
				scanner.Split(bufio.ScanLines)
				for scanner.Scan() {
					line := scanner.Text()
					output += "\n" + line
					log.Infof("[SERVICE:%s (stderr)> ] %s", service.Name, line)
				}
			}()
			go func() {
				scanner := bufio.NewScanner(outPipe)
				scanner.Split(bufio.ScanLines)
				for scanner.Scan() {
					line := scanner.Text()
					output += "\n" + line
					log.Infof("[SERVICE:%s (stdout)> ] %s", service.Name, line)
				}
			}()
			cmd.Wait()
			service.commandResponseChan <- &ServiceCommandOutput{
				Output:   output,
				ExitCode: cmd.ProcessState.ExitCode(),
			}
		}
	}()
}

func (service *Service) startFetchRoutine() {
	go func() {
		for {
			service.ExecCommandInServiceBaseFolder(&ServiceCommand{
				Cmd: *exec.Command("git", "fetch", "--all"),
			})
			// TODO: Maybe use https://github.com/go-git/go-git
			// service.ExecCommandInServiceFolder(&ServiceCommand{
			// 	Cmd: *exec.Command("git", "log", `--pretty=format:'{%n  "commit": "%H",%n  "abbreviated_commit": "%h",%n  "tree": "%T",%n  "abbreviated_tree": "%t",%n  "parent": "%P",%n  "abbreviated_parent": "%p",%n  "refs": "%D",%n  "encoding": "%e",%n  "subject": "%s",%n  "sanitized_subject_line": "%f",%n  "body": "%b",%n  "commit_notes": "%N",%n  "verification_flag": "%G?",%n  "signer": "%GS",%n  "signer_key": "%GK",%n  "author": {%n    "name": "%aN",%n    "email": "%aE",%n    "date": "%aD"%n  },%n  "commiter": {%n    "name": "%cN",%n    "email": "%cE",%n    "date": "%cD"%n  }%n},'`),
			// })
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
