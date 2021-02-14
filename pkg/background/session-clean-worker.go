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

			shouldTryCleanFolders := false
			if killReason == models.KillReasonBuildFailed || killReason == models.KillReasonHealthcheckFailed {
				maxRetries := session.Application.Startup.Retries
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
				if *session.Application.CleanOnExit {
					session.LogInfo(fmt.Sprintf("Deleting session folder %s", session.Folder))
					err := os.RemoveAll(session.Folder)
					if err != nil {
						session.LogError(fmt.Sprintf("Error while removing session folder: %s", err.Error()))
					}
				}
			}

		}
	}()
}
