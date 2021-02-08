package background

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	log "github.com/sirupsen/logrus"
	"github.com/wufe/polo/pkg/models"
	"github.com/wufe/polo/pkg/versioning"
)

type ApplicationFetchWorker struct {
	mediator *Mediator
}

func NewApplicationFetchWorker(mediator *Mediator) *ApplicationFetchWorker {
	worker := &ApplicationFetchWorker{
		mediator: mediator,
	}

	worker.startAcceptingFetchRequests()

	return worker
}

func (w *ApplicationFetchWorker) startAcceptingFetchRequests() {
	go func() {
		for {
			application := <-w.mediator.ApplicationFetch.RequestChan
			w.FetchApplicationRemote(application)
			w.mediator.ApplicationFetch.ResponseChan <- nil
		}
	}()
}

func (w *ApplicationFetchWorker) FetchApplicationRemote(application *models.Application) {

	// TODO: Handle all these errors

	auth, err := application.GetAuth()
	if err != nil {
		return
	}

	gitClient := versioning.GetGitClient(application, auth)

	// Open repository
	repo, err := git.PlainOpen(application.BaseFolder)
	defaultApplicationErrorLog(application, err)
	if err != nil {
		return
	}

	// Fetch
	err = gitClient.FetchAll(application.BaseFolder)
	defaultApplicationErrorLog(application, err, git.NoErrAlreadyUpToDate)

	// Branches
	branches, err := repo.Branches()
	defaultApplicationErrorLog(application, err)
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
	defaultApplicationErrorLog(application, err)

	// Tags
	tags, err := repo.Tags()
	if err != nil {
		return
	}
	defaultApplicationErrorLog(application, err)

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
	defaultApplicationErrorLog(application, err)
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

	log.Infof("[APP:%s] Found %d commits", application.Name, len(application.Commits))
}
