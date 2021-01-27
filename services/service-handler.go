package services

import (
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
	"github.com/wufe/polo/models"
	"github.com/wufe/polo/utils"
)

type ServiceHandler struct {
	configuration *models.RootConfiguration
}

func NewServiceHandler(rootConfiguration *models.RootConfiguration) *ServiceHandler {
	return &ServiceHandler{
		configuration: rootConfiguration,
	}
}

func (serviceHandler *ServiceHandler) InitializeService(service *models.Service) error {
	sessionsFolder, err := filepath.Abs(serviceHandler.configuration.Global.SessionsFolder)
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

	serviceHandler.startServiceCommandWatch(service)
	serviceHandler.fetchServiceRemote(service)
	serviceHandler.startServiceFetchRoutine(service)

	return nil
}

func (serviceHandler *ServiceHandler) startServiceCommandWatch(service *models.Service) {
	go func() {
		for {
			cmd := <-service.CommandChan

			output := []string{}

			err := utils.ThroughCallback(utils.ExecuteCommand(&cmd.Cmd))(func(line string) {
				output = append(output, line)
				log.Infof("[SERVICE:%s (stdout)> ] %s", service.Name, line)
			})

			if err != nil {
				log.Errorf("[SERVICE:%s] %s", service.Name, err.Error())
				return
			}

			service.CommandResponseChan <- &models.ServiceCommandOutput{
				Output:   output,
				ExitCode: cmd.ProcessState.ExitCode(),
			}
		}
	}()
}

func (serviceHandler *ServiceHandler) defaultServiceErrorLog(service *models.Service, err error, except ...error) {
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

func (serviceHandler *ServiceHandler) fetchServiceRemote(service *models.Service) {

	// Open repository
	repo, err := git.PlainOpen(service.ServiceBaseFolder)
	serviceHandler.defaultServiceErrorLog(service, err)
	if err != nil {
		return
	}

	// Fetch
	err = repo.Fetch(&git.FetchOptions{
		RefSpecs: []config.RefSpec{"refs/*:refs/*", "HEAD:refs/heads/HEAD"},
	})
	serviceHandler.defaultServiceErrorLog(service, err, git.NoErrAlreadyUpToDate)
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return
	}

	// Get remote
	remote, err := repo.Remote("origin")
	serviceHandler.defaultServiceErrorLog(service, err)
	if err != nil {
		return
	}

	// Branches
	refs, err := remote.List(&git.ListOptions{
		// Auth: &http.BasicAuth{
		// 	Username: ,
		// }
	})
	serviceHandler.defaultServiceErrorLog(service, err)

	refPrefix := "refs/heads/"
	for _, ref := range refs {
		refName := ref.Name().String()
		if !strings.HasPrefix(refName, refPrefix) {
			continue
		}
		branchName := refName[len(refPrefix):]

		appendWithoutDup(service.Branches, branchName)

		service.ObjectsToHashMap[branchName] = ref.Hash().String()
		service.ObjectsToHashMap[fmt.Sprintf("origin/%s", branchName)] = ref.Hash().String()
		service.ObjectsToHashMap[ref.Name().String()] = ref.Hash().String()

		if service.HashToObjectsMap[ref.Hash().String()] == nil {
			service.HashToObjectsMap[ref.Hash().String()] = &models.RemoteObject{
				Branches: []string{},
				Tags:     []string{},
			}
		}

		appendWithoutDup(service.HashToObjectsMap[ref.Hash().String()].Branches, branchName)
	}

	// Tags
	tags, err := repo.Tags()
	serviceHandler.defaultServiceErrorLog(service, err)

	tagPrefix := "refs/tags/"
	err = tags.ForEach(func(ref *plumbing.Reference) error {
		refName := ref.Name().String()
		if !strings.HasPrefix(refName, tagPrefix) {
			return nil
		}
		tagName := refName[len(tagPrefix):]
		service.ObjectsToHashMap[tagName] = ref.Hash().String()

		appendWithoutDup(service.Tags, tagName)
		service.ObjectsToHashMap[refName] = ref.Hash().String()

		if service.HashToObjectsMap[ref.Hash().String()] == nil {
			service.HashToObjectsMap[ref.Hash().String()] = &models.RemoteObject{
				Branches: []string{},
				Tags:     []string{},
			}
		}

		appendWithoutDup(service.HashToObjectsMap[ref.Hash().String()].Tags, tagName)

		return nil
	})

	// Log
	// TODO: Configure "since"
	since := time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC)
	until := time.Now().UTC()
	logs, err := repo.Log(&git.LogOptions{All: true, Since: &since, Until: &until, Order: git.LogOrderCommitterTime})
	serviceHandler.defaultServiceErrorLog(service, err)
	if err != nil {
		return
	}

	service.Commits = []string{}

	err = logs.ForEach(func(commit *object.Commit) error {
		service.ObjectsToHashMap[commit.Hash.String()] = commit.Hash.String()
		service.Commits = append(service.Commits, commit.Hash.String())
		service.CommitMap[commit.Hash.String()] = commit
		return nil
	})

	log.Infof("[SERVICE:%s] Found %d commits", service.Name, len(service.Commits))
}

func (serviceHandler *ServiceHandler) startServiceFetchRoutine(service *models.Service) {
	go func() {
		for {
			time.Sleep(5 * time.Minute)

			serviceHandler.fetchServiceRemote(service)
		}
	}()
}

func (serviceHandler *ServiceHandler) ExecServiceCommand(service *models.Service, command *models.ServiceCommand) *models.ServiceCommandOutput {
	service.CommandChan <- command
	return <-service.CommandResponseChan
}

func (serviceHandler *ServiceHandler) ExecServiceCommandInServiceFolder(service *models.Service, command *models.ServiceCommand) *models.ServiceCommandOutput {
	command.Dir = service.ServiceFolder
	return serviceHandler.ExecServiceCommand(service, command)
}

func (serviceHandler *ServiceHandler) ExecServiceCommandInServiceBaseFolder(service *models.Service, command *models.ServiceCommand) *models.ServiceCommandOutput {
	command.Dir = service.ServiceBaseFolder
	return serviceHandler.ExecServiceCommand(service, command)
}

func (serviceHandler *ServiceHandler) ExecServiceCommandInServiceCheckoutFolder(service *models.Service, command *models.ServiceCommand, checkout string) *models.ServiceCommandOutput {
	command.Dir = filepath.Join(service.ServiceFolder, checkout)
	return serviceHandler.ExecServiceCommand(service, command)
}
