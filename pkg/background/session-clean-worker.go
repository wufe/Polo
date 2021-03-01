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

	worker.startAcceptingSessionCleanRequests()

	return worker
}

func (w *SessionCleanWorker) startAcceptingSessionCleanRequests() {
	go func() {
		for {
			sessionToClean := <-w.mediator.CleanSession.Chan
			session := sessionToClean.Session
			session.LogInfo("Cleaning up session")
			session.SetStatus(sessionToClean.Status)
			w.sessionStorage.Delete(session)
			session.LogInfo("Session cleaned up")

			killReason := session.GetKillReason()
			conf := session.GetConfiguration()
			appStartupRetries := conf.Startup.Retries
			appCleanOnExit := *conf.CleanOnExit

			shouldTryCleanFolders := false
			if killReason == models.KillReasonBuildFailed || killReason == models.KillReasonHealthcheckFailed {
				maxRetries := appStartupRetries
				if maxRetries > 0 {
					retriesCount := session.GetStartupRetriesCount()
					if retriesCount < maxRetries {
						retriesCount++
						session.LogWarn(fmt.Sprintf("[%d/%d] Retrying session startup.", retriesCount, maxRetries))
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
					if replacement.Replaces() == session {
						// If so, tell this session that it is not a replacement anymore
						replacement.IsReplacementFor(nil)
						// And destroy it too
						replacement.SetKillReason(models.KillReasonStopped)
						w.mediator.DestroySession.Enqueue(replacement, nil)
						w.sessionStorage.Update(replacement)
					}
				}
			}

		}
	}()
}
