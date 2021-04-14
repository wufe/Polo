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

	var baseFolder string

	application.WithRLock(func(a *models.Application) {
		baseFolder = a.BaseFolder
	})

	conf := application.GetConfiguration()
	appName := conf.Name

	objectsToHashMap := make(map[string]string)
	hashToObjectsMap := make(map[string]*models.RemoteObject)
	appBranches := make(map[string]*models.Branch)
	appTags := make(map[string]*models.Tag)
	appCommits := []string{}
	appCommitMap := make(map[string]*object.Commit)

	checkObjectExists := func(hashToObjectsMap map[string]*models.RemoteObject) func(hash string) {
		return func(hash string) {
			if _, exists := hashToObjectsMap[hash]; !exists {
				hashToObjectsMap[hash] = &models.RemoteObject{
					Branches: []string{},
					Tags:     []string{},
				}
			}
		}
	}(hashToObjectsMap)

	aliveRefs := []string{}
	for _, s := range w.sessionStorage.GetAllAliveSessions() {
		s.RLock()
		ref := s.Checkout
		s.RUnlock()
		aliveRefs = append(aliveRefs, ref)
	}
	registerHash, watchResults := w.registerObjectHash(objectsToHashMap, conf.Branches, aliveRefs)

	gitClient := versioning.GetGitClient(application)

	// Open repository
	repo, err := git.PlainOpen(baseFolder)
	defaultApplicationErrorLog(appName, err)
	if err != nil {
		return
	}

	// Fetch
	err = gitClient.FetchAll(baseFolder)
	defaultApplicationErrorLog(appName, err, git.NoErrAlreadyUpToDate)

	// Branches
	branches, err := repo.Branches()
	defaultApplicationErrorLog(appName, err)
	refPrefix := "refs/heads/"
	branches.ForEach(func(ref *plumbing.Reference) error {
		refName := ref.Name().String()
		refHash := ref.Hash().String()

		if !strings.HasPrefix(refName, refPrefix) {
			return nil
		}
		branchName := refName[len(refPrefix):]

		registerHash(branchName, refHash)
		registerHash(fmt.Sprintf("origin/%s", branchName), refHash)
		registerHash(ref.Name().String(), refHash)

		checkObjectExists(refHash)

		hashToObjectsMap[refHash].Branches = appendWithoutDup(hashToObjectsMap[refHash].Branches, branchName)

		commit, err := repo.CommitObject(ref.Hash())
		if err != nil {
			return err
		}

		appBranches[branchName] = &models.Branch{
			CheckoutObject: models.CheckoutObject{
				Name:        branchName,
				Hash:        refHash,
				Author:      commit.Author.Name,
				AuthorEmail: commit.Author.Email,
				Date:        commit.Author.When,
				Message:     commit.Message,
			},
		}

		return nil
	})
	defaultApplicationErrorLog(appName, err)

	// Tags
	tags, err := repo.Tags()
	defaultApplicationErrorLog(appName, err)
	if err != nil {
		return
	}

	tagPrefix := "refs/tags/"
	err = tags.ForEach(func(ref *plumbing.Reference) error {
		refName := ref.Name().String()
		refHash := ref.Hash().String()

		if !strings.HasPrefix(refName, tagPrefix) {
			return nil
		}

		tagName := refName[len(tagPrefix):]
		registerHash(tagName, refHash)

		commit, err := repo.CommitObject(ref.Hash())
		if err != nil {
			return err
		}

		// appTags = appendWithoutDup(appTags, tagName)
		appTags[tagName] = &models.Tag{
			CheckoutObject: models.CheckoutObject{
				Name:        tagName,
				Hash:        refHash,
				Author:      commit.Author.Name,
				AuthorEmail: commit.Author.Email,
				Date:        commit.Author.When,
				Message:     commit.Message,
			},
		}
		registerHash(refName, refHash)
		checkObjectExists(refHash)

		hashToObjectsMap[refHash].Tags = appendWithoutDup(hashToObjectsMap[refHash].Tags, tagName)

		return nil
	})

	// Annotated tags
	tagObjects, err := repo.TagObjects()
	defaultApplicationErrorLog(appName, err)
	if err != nil {
		return
	}

	err = tagObjects.ForEach(func(ref *object.Tag) error {
		refName := ref.Name
		refHash := ref.Hash.String()

		tagName := refName

		registerHash(tagName, refHash)

		commit, err := ref.Commit()
		if err != nil {
			return err
		}

		// appTags = appendWithoutDup(appTags, tagName)
		appTags[tagName] = &models.Tag{
			CheckoutObject: models.CheckoutObject{
				Name:        tagName,
				Hash:        refHash,
				Author:      commit.Author.Name,
				AuthorEmail: commit.Author.Email,
				Date:        commit.Author.When,
				Message:     commit.Message,
			},
		}
		registerHash(refName, refHash)
		checkObjectExists(refHash)

		hashToObjectsMap[refHash].Tags = appendWithoutDup(hashToObjectsMap[refHash].Tags, tagName)

		appCommitMap[refHash] = commit

		return nil
	})

	// Log
	// TODO: Configure "since"
	since := time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC)
	until := time.Now().UTC()
	logs, err := repo.Log(&git.LogOptions{All: true, Since: &since, Until: &until, Order: git.LogOrderCommitterTime})
	defaultApplicationErrorLog(appName, err)
	if err != nil {
		return
	}

	err = logs.ForEach(func(commit *object.Commit) error {
		commitHash := commit.Hash.String()
		registerHash(commitHash, commitHash)
		appCommits = append(appCommits, commitHash)
		appCommitMap[commitHash] = commit
		return nil
	})

	var lastCommitsCount int
	newCommitsCount := len(appCommits)
	application.WithLock(func(a *models.Application) {
		lastCommitsCount = len(a.Commits)

		a.ObjectsToHashMap = objectsToHashMap
		a.HashToObjectsMap = hashToObjectsMap
		a.BranchesMap = appBranches
		a.TagsMap = appTags
		a.Commits = appCommits
		a.CommitMap = appCommitMap
	})

	if newCommitsCount > lastCommitsCount {
		log.Infof("[APP:%s] Found %d new commits", appName, newCommitsCount-lastCommitsCount)
	}

	if !watchObjects {
		return
	}

	for ref, hash := range *watchResults {
		sessions := w.sessionStorage.GetAllAliveSessions()
		var foundSession *models.Session
		for _, session := range sessions {

			sessionAppName := session.ApplicationName
			sessionCheckout := session.Checkout

			if sessionAppName == appName && (sessionCheckout == ref) {
				foundSession = session
			}
		}
		buildSession := requestSessionBuilder(application, ref)
		if foundSession != nil {
			sessionCommitID := foundSession.CommitID
			if sessionCommitID != hash {
				log.Infof("[APP:%s][WATCH] Detected new commit on %s", appName, ref)
				// FEATURE: Hot swap
				// Set the previous' session kill-reason to "replaced"
				// and create a new session.
				// This new one will be aware that it is a replacement for another session that is going to expire.
				// When the new one gets started, the old one gets destroyed.
				foundSession.SetKillReason(models.KillReasonReplaced)
				buildSession(w.mediator, foundSession)
			}
		} else {

			var lastSession *models.Session
			allSessions := w.sessionStorage.GetByApplicationName(appName)
			if len(allSessions) > 0 {
				for _, s := range allSessions {
					if s.Checkout == ref {
						lastSession = s
					}
				}
			}

			if lastSession == nil || !lastSession.GetKillReason().PreventsRebuild() {
				log.Infof("[APP:%s][WATCH] Auto-start on %s", appName, ref)
				buildSession(w.mediator, nil)
			}
		}
	}
}

func (w *ApplicationFetchWorker) registerObjectHash(objectsToHashMap map[string]string, branches models.Branches, aliveRefs []string) (func(refName string, hash string), *map[string]string) {
	watchResults := make(map[string]string)
	watchedHashes := make(map[string]bool)

	aliveContains := func(ref string) bool {
		for _, r := range aliveRefs {
			if r == ref {
				return true
			}
		}
		return false
	}

	return func(refName, hash string) {
		objectsToHashMap[refName] = hash
		// If the branch is being watched or the branch is being served from an alive session
		if branches.BranchIsBeingWatched(refName) || aliveContains(refName) {
			if _, ok := watchedHashes[hash]; !ok {
				watchResults[refName] = hash
				watchedHashes[hash] = true
			}
		}
	}, &watchResults
}

func requestSessionBuilder(a *models.Application, ref string) func(*Mediator, *models.Session) {
	return func(mediator *Mediator, previousSession *models.Session) {
		mediator.BuildSession.Enqueue(ref, a, previousSession)
	}
}

func defaultApplicationErrorLog(name string, err error, except ...error) {
	if err != nil {
		var foundError error
		for _, exceptErr := range except {
			if exceptErr == err {
				foundError = exceptErr
			}
		}
		if foundError == nil {
			log.Errorf("[APP:%s] %s", name, err.Error())
		}
	}
}
