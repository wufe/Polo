package background

import (
	log "github.com/sirupsen/logrus"
)

type SessionCleanWorker struct {
	mediator *Mediator
}

func NewSessionCleanWorker(mediator *Mediator) *SessionCleanWorker {
	worker := &SessionCleanWorker{
		mediator: mediator,
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
			log.Warnf("[SESSION:%s] Session cleaned up.", sessionToClean.Session.UUID)
		}
	}()
}
