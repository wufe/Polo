package background

import (
	"time"

	"github.com/wufe/polo/pkg/logging"
	"github.com/wufe/polo/pkg/models"
	"github.com/wufe/polo/pkg/storage"
)

type SessionStartWorker struct {
	sessionStorage *storage.Session
	mediator       *Mediator
	logger         logging.Logger
}

func NewSessionStartWorker(
	sessionStorage *storage.Session,
	mediator *Mediator,
	logger logging.Logger,
) *SessionStartWorker {
	worker := &SessionStartWorker{
		sessionStorage: sessionStorage,
		mediator:       mediator,
		logger:         logger,
	}
	return worker
}

func (w *SessionStartWorker) Start() {
	w.startAcceptingSessionStartRequests()
}

func (w *SessionStartWorker) startAcceptingSessionStartRequests() {
	go func() {
		for {
			request := <-w.mediator.StartSession.Chan
			w.MarkSessionAsStarted(request.Session)
		}
	}()
}

func (w *SessionStartWorker) MarkSessionAsStarted(session *models.Session) {
	session.SetStatus(models.SessionStatusStarted)
	session.ResetStartupRetriesCount()
	conf := session.GetConfiguration()
	if conf.Branches.BranchIsBeingWatched(session.Checkout, w.logger) {
		session.SetMaxAge(-1)
	} else {
		session.SetMaxAge(conf.Recycle.InactivityTimeout)
		if session.GetMaxAge() > 0 {
			w.startSessionInactivityTimer(session)
		}
	}

	// FEATURE: Hot swap
	// Checks if this session replaces something else
	replaces := session.GetReplaces()
	if len(replaces) > 0 {
		for _, replaced := range replaces {
			// Notify the previous one that it has been replaced
			replaced.SetReplacedBy(session)
			// And destroy it
			w.mediator.DestroySession.Enqueue(replaced, nil)
		}
	}
	// Reset status of current session
	session.SetReplaces(nil)

	w.sessionStorage.Update(session)

	session.GetEventBus().PublishEvent(models.SessionEventTypeSessionStarted, session)
}

func (w *SessionStartWorker) startSessionInactivityTimer(session *models.Session) {
	conf := session.GetConfiguration()
	session.SetInactiveAt(time.Now().Add(time.Second * time.Duration(conf.Recycle.InactivityTimeout)))
	go func() {
		for {
			status := session.GetStatus()
			if status != models.SessionStatusStarted {
				return
			}

			if time.Now().After(session.GetInactiveAt()) {
				w.mediator.DestroySession.Enqueue(session, nil)
				return
			}
			session.DecreaseMaxAge()
			time.Sleep(1 * time.Second)
		}
	}()
}
