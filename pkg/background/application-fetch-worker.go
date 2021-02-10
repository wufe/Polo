package background

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	log "github.com/sirupsen/logrus"
	"github.com/wufe/polo/pkg/background/pipe"
	"github.com/wufe/polo/pkg/models"
	"github.com/wufe/polo/pkg/storage"
	"github.com/wufe/polo/pkg/versioning"
)

type ApplicationFetchWorker struct {
	sessionStorage *storage.Session
	mediator       *Mediator
}

func NewApplicationFetchWorker(sessionStorage *storage.Session, mediator *Mediator) *ApplicationFetchWorker {
	worker := &ApplicationFetchWorker{
		sessionStorage: sessionStorage,
		mediator:       mediator,
	}

	worker.startAcceptingFetchRequests()

	return worker
}

func (w *ApplicationFetchWorker) startAcceptingFetchRequests() {
	go func() {
		for {
			applicationFetchReq := <-w.mediator.ApplicationFetch.RequestChan
			w.FetchApplicationRemote(applicationFetchReq.Application, applicationFetchReq.WatchObjects)
			w.mediator.ApplicationFetch.ResponseChan <- nil
		}
	}()
}

func (w *ApplicationFetchWorker) FetchApplicationRemote(application *models.Application, watchObjects bool) {

	// TODO: Handle all these errors

	registerHash, watchResults := w.registerObjectHash(application)

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

		registerHash(branchName, ref.Hash().String())
		registerHash(fmt.Sprintf("origin/%s", branchName), ref.Hash().String())
		registerHash(ref.Name().String(), ref.Hash().String())

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
		registerHash(tagName, ref.Hash().String())

		application.Tags = appendWithoutDup(application.Tags, tagName)
		registerHash(refName, ref.Hash().String())

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
		registerHash(commit.Hash.String(), commit.Hash.String())
		application.Commits = append(application.Commits, commit.Hash.String())
		application.CommitMap[commit.Hash.String()] = commit
		return nil
	})

	log.Infof("[APP:%s] Found %d commits", application.Name, len(application.Commits))

	if !watchObjects {
		return
	}

	for ref, hash := range *watchResults {
		sessions := w.sessionStorage.GetAllAliveSessions()
		var foundSession *models.Session
		for _, session := range sessions {
			if session.Application == application && (session.Checkout == ref) {
				foundSession = session
			}
		}
		buildSession := requestSessionBuilder(application, ref)
		if foundSession != nil {
			if foundSession.CommitID != hash {
				log.Infof("[APP:%s][WATCH] Detected new commit on %s", application.Name, ref)
				w.mediator.DestroySession.Request(foundSession, func(s *models.Session) {
					buildSession(w.mediator)
				})
			}
		} else {
			log.Infof("[APP:%s][WATCH] Auto-start on %s", application.Name, ref)
			buildSession(w.mediator)
		}
	}
}

func (w *ApplicationFetchWorker) registerObjectHash(a *models.Application) (func(refName string, hash string), *map[string]string) {
	watchResults := make(map[string]string)
	watchedHashes := make(map[string]bool)
	return func(refName, hash string) {
		a.ObjectsToHashMap[refName] = hash
		if a.Watch.Contains(refName) {
			if _, ok := watchedHashes[hash]; !ok {
				watchResults[refName] = hash
				watchedHashes[hash] = true
			}
		}
	}, &watchResults
}

func requestSessionBuilder(a *models.Application, ref string) func(*Mediator) {
	return func(mediator *Mediator) {
		mediator.BuildSession.Request(&pipe.SessionBuildInput{
			Application: a,
			Checkout:    ref,
		})
	}
}
