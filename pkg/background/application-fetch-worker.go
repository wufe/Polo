package background

import (
	log "github.com/sirupsen/logrus"
	"github.com/wufe/polo/pkg/models"
	"github.com/wufe/polo/pkg/storage"
	"github.com/wufe/polo/pkg/versioning"
)

type ApplicationFetchWorker struct {
	sessionStorage    *storage.Session
	mediator          *Mediator
	repositoryFetcher versioning.RepositoryFetcher
}

func NewApplicationFetchWorker(sessionStorage *storage.Session, repositoryFetcher versioning.RepositoryFetcher, mediator *Mediator) *ApplicationFetchWorker {
	worker := &ApplicationFetchWorker{
		sessionStorage:    sessionStorage,
		repositoryFetcher: repositoryFetcher,
		mediator:          mediator,
	}
	return worker
}

func (w *ApplicationFetchWorker) Start() {
	w.startAcceptingFetchRequests()
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

	bus := application.GetEventBus()
	bus.PublishEvent(models.ApplicationEventTypeFetchStarted, application)
	defer bus.PublishEvent(models.ApplicationEventTypeFetchCompleted, application)

	var baseFolder string

	application.WithRLock(func(a *models.Application) {
		baseFolder = a.BaseFolder
	})

	conf := application.GetConfiguration()
	appName := conf.Name

	fetchResult, errors := w.repositoryFetcher.Fetch(baseFolder)
	if len(errors) > 0 {
		for _, err := range errors {
			defaultApplicationErrorLog(appName, err)
		}
	}

	// Something gone terribly wrong
	if fetchResult == nil {
		return
	}

	var lastCommitsCount int
	newCommitsCount := len(fetchResult.AppCommits)
	application.WithLock(func(a *models.Application) {
		lastCommitsCount = len(a.Commits)

		a.ObjectsToHashMap = fetchResult.ObjectsToHashMap
		a.HashToObjectsMap = fetchResult.HashToObjectsMap
		a.BranchesMap = fetchResult.BranchesMap
		a.TagsMap = fetchResult.TagsMap
		a.Commits = fetchResult.AppCommits
		a.CommitMap = fetchResult.CommitMap
	})

	if newCommitsCount > lastCommitsCount {
		log.Infof("[APP:%s] Found %d new commits", appName, newCommitsCount-lastCommitsCount)
	}

	if !watchObjects {
		return
	}

	aliveRefs := []string{}
	for _, s := range w.sessionStorage.GetAllAliveSessions() {
		s.RLock()
		ref := s.Checkout
		s.RUnlock()
		aliveRefs = append(aliveRefs, ref)
	}

	watchResults := make(map[string]string)
	watchedHashes := make(map[string]bool)
	branches := conf.Branches

	aliveContains := func(ref string) bool {
		for _, r := range aliveRefs {
			if r == ref {
				return true
			}
		}
		return false
	}

	for refName, hash := range fetchResult.ObjectsToHashMap {
		if branches.BranchIsBeingWatched(refName) || aliveContains(refName) {
			if _, exists := watchedHashes[hash]; !exists {
				watchedHashes[hash] = true
				watchResults[refName] = hash
			}
		}
	}

	for ref, hash := range watchResults {
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
				bus.PublishEvent(models.ApplicationEventTypeHotSwap, application, foundSession)
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
				bus.PublishEvent(models.ApplicationEventTypeAutoStart, application)
				buildSession(w.mediator, nil)
			}
		}
	}
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
