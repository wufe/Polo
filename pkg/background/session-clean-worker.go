package background

import (
	log "github.com/sirupsen/logrus"
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
			sessionToClean.Session.LogInfo("Cleaning up session")
			sessionToClean.Session.Status = sessionToClean.Status
			w.sessionStorage.Delete(sessionToClean.Session)
			log.Warnf("[SESSION:%s] Session cleaned up.", sessionToClean.Session.UUID)
		}
	}()
}
