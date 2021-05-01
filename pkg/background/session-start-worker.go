package background

import (
	"time"

	"github.com/wufe/polo/pkg/models"
	"github.com/wufe/polo/pkg/storage"
)

type SessionStartWorker struct {
	sessionStorage *storage.Session
	mediator       *Mediator
}

func NewSessionStartWorker(
	sessionStorage *storage.Session,
	mediator *Mediator,
) *SessionStartWorker {
	worker := &SessionStartWorker{
		sessionStorage: sessionStorage,
		mediator:       mediator,
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
	session.GetEventBus().PublishEvent(models.SessionBuildEventTypeStarted, session)
	session.SetStatus(models.SessionStatusStarted)
	session.ResetStartupRetriesCount()
	conf := session.GetConfiguration()
	if conf.Branches.BranchIsBeingWatched(session.Checkout) {
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
	// If replaces another session and that session has been killed for replacements reason
	if replaces != nil && replaces.GetKillReason() == models.KillReasonReplaced {
		// Notify the previous one that it has been replaced
		replaces.SetReplacedBy(session.UUID)
		// And destroy it
		w.mediator.DestroySession.Enqueue(replaces, nil)
	}
	// Reset status of current session
	session.SetReplaces(nil)

	w.sessionStorage.Update(session)
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
