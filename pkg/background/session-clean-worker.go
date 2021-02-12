package background

import (
	"fmt"

	log "github.com/sirupsen/logrus"
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
			log.Warnf("[SESSION:%s] Session cleaned up.", session.UUID)

			killReason := session.GetKillReason()

			if killReason == models.KillReasonBuildFailed || killReason == models.KillReasonHealthcheckFailed {
				maxRetries := session.Application.Startup.Retries
				if maxRetries > 0 {
					retriesCount := session.GetStartupRetriesCount()
					if retriesCount < maxRetries {
						retriesCount++
						session.LogWarn(fmt.Sprintf("[%d/%d] Retrying session startup.", retriesCount, maxRetries))
						log.Warnf("[SESSION:%s][%d/%d] Retrying session startup.", session.UUID, retriesCount, maxRetries)
						w.mediator.BuildSession.Enqueue(session.Checkout, session.Application, session)
					} else {
						session.LogError("Max startup retries exceeded.")
						log.Warnf("[SESSION:%s] Session's max startup retries exceeded.", session.UUID)
					}
				}
			}

		}
	}()
}
