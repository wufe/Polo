package services

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/kennygrant/sanitize"
	log "github.com/sirupsen/logrus"
	"github.com/wufe/polo/models"
	"github.com/wufe/polo/services/versioning"
	"github.com/wufe/polo/utils"
)

type ApplicationHandler struct {
	configuration *models.RootConfiguration
}

func NewApplicationHandler(rootConfiguration *models.RootConfiguration) *ApplicationHandler {
	return &ApplicationHandler{
		configuration: rootConfiguration,
	}
}

func (applicationHandler *ApplicationHandler) InitializeApplication(application *models.Application) error {
	log.Infof("[APPLICATION:%s] Initializing", application.Name)
	sessionsFolder, err := filepath.Abs(applicationHandler.configuration.Global.SessionsFolder)
	if err != nil {
		return err
	}
	if _, err := os.Stat(sessionsFolder); os.IsNotExist(err) { // Session folder does not exist
		err := os.Mkdir(sessionsFolder, 0755)
		if err != nil {
			return err
		}
	}
	applicationName := sanitize.Name(application.Name)
	applicationFolder := filepath.Join(sessionsFolder, applicationName)
	if _, err := os.Stat(applicationFolder); os.IsNotExist(err) { // Application folder does not exist
		err := os.Mkdir(applicationFolder, 0755)
		if err != nil {
			return err
		}
	}
	application.Folder = applicationFolder

	baseFolder := filepath.Join(applicationFolder, "_base") // Folder used for performing periodic git fetch --all and/or git log
	if _, err := os.Stat(baseFolder); os.IsNotExist(err) {  // Application folder does not exist

		auth, err := application.GetAuth()
		if err != nil {
			return err
		}

		gitClient := versioning.GetGitClient(application, auth)

		err = gitClient.Clone(applicationFolder, "_base", application.Remote)
		if err != nil {
			return err
		}

	}
	application.BaseFolder = baseFolder

	applicationHandler.startApplicationCommandWatch(application)
	applicationHandler.fetchApplicationRemote(application)
	applicationHandler.startApplicationFetchRoutine(application)

	return nil
}

func (applicationHandler *ApplicationHandler) startApplicationCommandWatch(application *models.Application) {
	go func() {
		for {
			cmd := <-application.CommandChan

			output := []string{}

			err := utils.ExecCmds(func(sl *utils.StdLine) {
				output = append(output, sl.Line)
				log.Infof("[APPLICATION:%s (stdout)> ] %s", application.Name, sl.Line)
			}, &cmd.Cmd)

			if err != nil {
				log.Errorf("[APPLICATION:%s] %s", application.Name, err.Error())
				return
			}

			application.CommandResponseChan <- &models.ApplicationCommandOutput{
				Output:   output,
				ExitCode: cmd.ProcessState.ExitCode(),
			}
		}
	}()
}

func (applicationHandler *ApplicationHandler) defaultApplicationErrorLog(application *models.Application, err error, except ...error) {
	if err != nil {
		var foundError error
		for _, exceptErr := range except {
			if exceptErr == err {
				foundError = exceptErr
			}
		}
		if foundError == nil {
			log.Errorf("[APPLICATION:%s] %s", application.Name, err.Error())
		}
	}
}

func appendWithoutDup(slice []string, elem ...string) []string {
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
	return slice
}

func (applicationHandler *ApplicationHandler) fetchApplicationRemote(application *models.Application) {

	// TODO: Handle all these errors

	auth, err := application.GetAuth()
	if err != nil {
		return
	}

	gitClient := versioning.GetGitClient(application, auth)

	// Open repository
	repo, err := git.PlainOpen(application.BaseFolder)
	applicationHandler.defaultApplicationErrorLog(application, err)
	if err != nil {
		return
	}

	// Fetch
	err = gitClient.FetchAll(application.BaseFolder)
	applicationHandler.defaultApplicationErrorLog(application, err, git.NoErrAlreadyUpToDate)

	// Branches
	branches, err := repo.Branches()
	applicationHandler.defaultApplicationErrorLog(application, err)
	refPrefix := "refs/heads/"
	application.Branches = make(map[string]*models.Branch)
	branches.ForEach(func(ref *plumbing.Reference) error {
		refName := ref.Name().String()
		if !strings.HasPrefix(refName, refPrefix) {
			return nil
		}
		branchName := refName[len(refPrefix):]

		application.ObjectsToHashMap[branchName] = ref.Hash().String()
		application.ObjectsToHashMap[fmt.Sprintf("origin/%s", branchName)] = ref.Hash().String()
		application.ObjectsToHashMap[ref.Name().String()] = ref.Hash().String()

		if application.HashToObjectsMap[ref.Hash().String()] == nil {
			application.HashToObjectsMap[ref.Hash().String()] = &models.RemoteObject{
				Branches: []string{},
				Tags:     []string{},
			}
		}

		application.HashToObjectsMap[ref.Hash().String()].Branches = appendWithoutDup(application.HashToObjectsMap[ref.Hash().String()].Branches, branchName)

		commit, err := repo.CommitObject(ref.Hash())
		if err != nil {
			return err
		}

		application.Branches[branchName] = &models.Branch{
			Name:    branchName,
			Hash:    ref.Hash().String(),
			Author:  commit.Author.Email,
			Date:    commit.Author.When,
			Message: commit.Message,
		}

		return nil
	})
	applicationHandler.defaultApplicationErrorLog(application, err)

	// Tags
	tags, err := repo.Tags()
	if err != nil {
		return
	}
	applicationHandler.defaultApplicationErrorLog(application, err)

	tagPrefix := "refs/tags/"
	err = tags.ForEach(func(ref *plumbing.Reference) error {
		refName := ref.Name().String()
		if !strings.HasPrefix(refName, tagPrefix) {
			return nil
		}
		tagName := refName[len(tagPrefix):]
		application.ObjectsToHashMap[tagName] = ref.Hash().String()

		application.Tags = appendWithoutDup(application.Tags, tagName)
		application.ObjectsToHashMap[refName] = ref.Hash().String()

		if application.HashToObjectsMap[ref.Hash().String()] == nil {
			application.HashToObjectsMap[ref.Hash().String()] = &models.RemoteObject{
				Branches: []string{},
				Tags:     []string{},
			}
		}

		application.HashToObjectsMap[ref.Hash().String()].Tags = appendWithoutDup(application.HashToObjectsMap[ref.Hash().String()].Tags, tagName)

		return nil
	})

	// Log
	// TODO: Configure "since"
	since := time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC)
	until := time.Now().UTC()
	logs, err := repo.Log(&git.LogOptions{All: true, Since: &since, Until: &until, Order: git.LogOrderCommitterTime})
	applicationHandler.defaultApplicationErrorLog(application, err)
	if err != nil {
		return
	}

	application.Commits = []string{}

	err = logs.ForEach(func(commit *object.Commit) error {
		application.ObjectsToHashMap[commit.Hash.String()] = commit.Hash.String()
		application.Commits = append(application.Commits, commit.Hash.String())
		application.CommitMap[commit.Hash.String()] = commit
		return nil
	})

	log.Infof("[APPLICATION:%s] Found %d commits", application.Name, len(application.Commits))
}

func (applicationHandler *ApplicationHandler) startApplicationFetchRoutine(application *models.Application) {
	go func() {
		for {
			time.Sleep(5 * time.Minute)

			applicationHandler.fetchApplicationRemote(application)
		}
	}()
}

func (applicationHandler *ApplicationHandler) ExecApplicationCommand(application *models.Application, command *models.ApplicationCommand) *models.ApplicationCommandOutput {
	application.CommandChan <- command
	return <-application.CommandResponseChan
}

func (applicationHandler *ApplicationHandler) ExecApplicationCommandInApplicationFolder(application *models.Application, command *models.ApplicationCommand) *models.ApplicationCommandOutput {
	command.Dir = application.Folder
	return applicationHandler.ExecApplicationCommand(application, command)
}

func (applicationHandler *ApplicationHandler) ExecApplicationCommandInApplicationBaseFolder(application *models.Application, command *models.ApplicationCommand) *models.ApplicationCommandOutput {
	command.Dir = application.BaseFolder
	return applicationHandler.ExecApplicationCommand(application, command)
}

func (applicationHandler *ApplicationHandler) ExecApplicationCommandInApplicationCheckoutFolder(application *models.Application, command *models.ApplicationCommand, checkout string) *models.ApplicationCommandOutput {
	command.Dir = filepath.Join(application.Folder, checkout)
	return applicationHandler.ExecApplicationCommand(application, command)
}
