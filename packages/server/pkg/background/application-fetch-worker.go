package background

import (
	"github.com/wufe/polo/pkg/logging"
	"github.com/wufe/polo/pkg/models"
	"github.com/wufe/polo/pkg/storage"
	"github.com/wufe/polo/pkg/versioning"
)

type ApplicationFetchWorker struct {
	sessionStorage    *storage.Session
	mediator          *Mediator
	repositoryFetcher versioning.RepositoryFetcher
	log               logging.Logger
}

func NewApplicationFetchWorker(sessionStorage *storage.Session, repositoryFetcher versioning.RepositoryFetcher, mediator *Mediator, log logging.Logger) *ApplicationFetchWorker {
	worker := &ApplicationFetchWorker{
		sessionStorage:    sessionStorage,
		repositoryFetcher: repositoryFetcher,
		mediator:          mediator,
		log:               log,
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
	appID := conf.ID

	fetchResult, errors := w.repositoryFetcher.Fetch(baseFolder)
	if len(errors) > 0 {
		for _, err := range errors {
			w.log.Errorf("Error while loading application: %s", err.Error.Error())
			w.defaultApplicationErrorLog(appName, err.Error)
			if err.Critical {
				application.AddNotification(
					models.ApplicationNotificationTypeGitFetch,
					err.Error.Error(),
					models.ApplicationNotificationLevelCritical,
					true,
				)
			}
		}
	} else {
		application.RemoveNotificationByType(models.ApplicationNotificationTypeGitFetch)
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
		w.log.Infof("[APP:%s] Found %d new commits", appName, newCommitsCount-lastCommitsCount)
	}

	if !watchObjects {
		return
	}

	aliveRefs := []string{}
	for _, s := range w.sessionStorage.GetAllAliveApplicationSessions(appID) {
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
		if branches.BranchIsBeingWatched(refName, w.log) || aliveContains(refName) {
			if _, exists := watchedHashes[hash]; !exists {
				watchedHashes[hash] = true
				watchResults[refName] = hash
			}
		}
	}

	for checkout, hash := range watchResults {
		lastSession := w.getLastAppSessionByCheckout(appName, checkout)
		buildSession := requestSessionBuilder(application, checkout)
		if lastSession != nil {
			sessionCommitID := lastSession.CommitID
			if sessionCommitID != hash {
				w.log.Infof("[APP:%s][WATCH] Detected new commit on %s", appName, checkout)
				// FEATURE: Hot swap
				// Set the previous' session kill-reason to "replaced"
				// and create a new session.
				// This new one will be aware that it is a replacement for another session that is going to expire.
				// When the new one gets started, the old one gets destroyed.
				bus.PublishEvent(models.ApplicationEventTypeHotSwap, application, lastSession)
				buildSession(w.mediator, nil, w.getAllAppSessionsToBeReplaced(appID, appName, checkout))
			}
		} else {

			var lastSession *models.Session
			allSessions := w.sessionStorage.GetByApplicationName(appName)
			if len(allSessions) > 0 {
				for _, s := range allSessions {
					if s.Checkout == checkout {
						lastSession = s
					}
				}
			}

			if lastSession == nil || !lastSession.GetKillReason().PreventsRebuild() {
				w.log.Infof("[APP:%s][WATCH] Auto-start on %s", appName, checkout)
				bus.PublishEvent(models.ApplicationEventTypeAutoStart, application)
				buildSession(w.mediator, nil, w.getAllAppSessionsToBeReplaced(appID, appName, checkout))
			}
		}
	}
}

func (w *ApplicationFetchWorker) getLastAppSessionByCheckout(appName string, checkout string) *models.Session {
	sessions := w.sessionStorage.GetByApplicationName(appName)
	var foundSession *models.Session
	for _, session := range sessions {
		sessionCheckout := session.Checkout

		if sessionCheckout == checkout {
			foundSession = session
		}
	}
	return foundSession
}

func (w *ApplicationFetchWorker) getAllAppSessionsToBeReplaced(appID, appName string, checkout string) []*models.Session {
	foundSessions := []*models.Session{}
	sessions := w.sessionStorage.GetAllAliveApplicationSessions(appID)
	for _, session := range sessions {
		sessionCheckout := session.Checkout
		replacedBy := session.GetReplacedBy()
		if sessionCheckout == checkout && replacedBy == nil {
			foundSessions = append(foundSessions, session)
		}
	}
	return foundSessions
}

func requestSessionBuilder(a *models.Application, ref string) func(*Mediator, *models.Session, []*models.Session) {
	return func(mediator *Mediator, previousSession *models.Session, sessionsToBeReplaced []*models.Session) {
		mediator.BuildSession.Enqueue(ref, a, previousSession, sessionsToBeReplaced, false)
	}
}

func (w *ApplicationFetchWorker) defaultApplicationErrorLog(name string, err error, except ...error) {
	if err != nil {
		var foundError error
		for _, exceptErr := range except {
			if exceptErr == err {
				foundError = exceptErr
			}
		}
		if foundError == nil {
			w.log.Errorf("[APP:%s] %s", name, err.Error())
		}
	}
}
