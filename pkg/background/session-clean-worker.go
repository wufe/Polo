package background

import (
	"fmt"
	"os"

	"github.com/wufe/polo/pkg/models"
	"github.com/wufe/polo/pkg/storage"
)

type SessionCleanWorker struct {
	sessionStorage *storage.Session
	mediator       *Mediator
}

func NewSessionCleanWorker(sessionStorage *storage.Session, mediator *Mediator) *SessionCleanWorker {
	worker := &SessionCleanWorker{
		sessionStorage: sessionStorage,
		mediator:       mediator,
	}
	return worker
}

func (w *SessionCleanWorker) Start() {
	w.startAcceptingSessionCleanRequests()
}

func (w *SessionCleanWorker) startAcceptingSessionCleanRequests() {
	go func() {
		for {
			sessionToClean := <-w.mediator.CleanSession.Chan
			session := sessionToClean.Session

			bus := session.GetEventBus()

			killReason := session.GetKillReason()
			session.LogInfo("Cleaning up session")
			session.SetStatus(sessionToClean.Status)
			// Even though this session is going to be deleted,
			// we are going to persist the status change
			w.sessionStorage.Update(sessionToClean.Session)
			w.sessionStorage.Delete(session)
			session.LogInfo("Session cleaned up")
			conf := session.GetConfiguration()
			appStartupRetries := conf.Startup.Retries
			appCleanOnExit := *conf.CleanOnExit

			shouldTryCleanFolders := false
			sessionGetsRecycled := false
			if killReason == models.KillReasonBuildFailed || killReason == models.KillReasonHealthcheckFailed {
				maxRetries := appStartupRetries
				if maxRetries > 0 {
					retriesCount := session.GetStartupRetriesCount()
					if retriesCount < maxRetries {
						sessionGetsRecycled = true
						bus.PublishEvent(models.SessionEventTypeGettingRecycled, session)
						retriesCount++
						session.LogWarn(fmt.Sprintf("[%d/%d] Retrying session startup.", retriesCount, maxRetries))
						bus.PublishEvent(models.SessionEventTypeBuildGettingRetried, session)
						w.mediator.BuildSession.Enqueue(session.Checkout, session.Application, session)
					} else {
						shouldTryCleanFolders = true
						session.LogWarn("Max startup retries exceeded.")
					}
				} else {
					shouldTryCleanFolders = true
				}
			} else {
				shouldTryCleanFolders = true
			}

			if shouldTryCleanFolders {
				if appCleanOnExit {
					bus.PublishEvent(models.SessionEventTypeFolderClean, session)
					session.LogInfo(fmt.Sprintf("Deleting session folder %s", session.Folder))
					err := os.RemoveAll(session.Folder)
					if err != nil {
						session.LogError(fmt.Sprintf("Error while removing session folder: %s", err.Error()))
					}
				}
			}

			if session.GetKillReason() == models.KillReasonStopped {
				// FEATURE: Hot swap
				// Check if the killed session should have been replaced by another session
				for _, replacement := range w.sessionStorage.GetAllAliveSessions() {
					if replacement.GetReplaces() == session {
						// If so, tell this session that it is not a replacement anymore
						replacement.SetReplaces(nil)
						// And destroy it too
						replacement.SetKillReason(models.KillReasonStopped)
						w.mediator.DestroySession.Enqueue(replacement, nil)
						w.sessionStorage.Update(replacement)
					}
				}
			}

			if !sessionGetsRecycled {
				bus := session.GetEventBus()
				bus.Close()
			}
		}
	}()
}
