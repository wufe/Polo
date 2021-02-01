package services

import (
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/wufe/polo/models"
)

type SessionHandler struct {
	configuration       *models.RootConfiguration
	serviceHandler      *ServiceHandler
	sessions            []*models.Session
	sessionRequestChan  chan *SessionBuildInput
	sessionResponseChan chan *SessionBuildResult
	sessionCleanChan    chan *SessionClean
}

type SessionClean struct {
	session *models.Session
	status  models.SessionStatus
}

func NewSessionHandler(configuration *models.RootConfiguration, serviceHandler *ServiceHandler) *SessionHandler {
	sessionHandler := &SessionHandler{
		configuration:       configuration,
		serviceHandler:      serviceHandler,
		sessions:            []*models.Session{},
		sessionRequestChan:  make(chan *SessionBuildInput),
		sessionResponseChan: make(chan *SessionBuildResult),
		sessionCleanChan:    make(chan *SessionClean),
	}

	sessionHandler.startAcceptingNewSessionRequests()
	sessionHandler.startAcceptingSessionCleanRequests()

	return sessionHandler
}

func (sessionHandler *SessionHandler) RequestNewSession(buildInput *SessionBuildInput) *SessionBuildResult {
	sessionHandler.sessionRequestChan <- buildInput
	buildResult := <-sessionHandler.sessionResponseChan
	return buildResult
}

func (sessionHandler *SessionHandler) GetSessionByUUID(uuid string) *models.Session {
	var foundSession *models.Session
	for _, session := range sessionHandler.sessions {
		if session.UUID == uuid {
			foundSession = session
		}
	}
	return foundSession
}

func (sessionHandler *SessionHandler) GetAllAliveSessions() []*models.Session {
	filteredSessions := []*models.Session{}
	for _, session := range sessionHandler.sessions {
		if session.Status.IsAlive() {
			filteredSessions = append(filteredSessions, session)
		}
	}
	return filteredSessions
}

func (sessionHandler *SessionHandler) GetAliveServiceSessionByCheckout(checkout string, service *models.Service) *models.Session {
	var foundSession *models.Session
	for _, session := range sessionHandler.sessions {
		if session.Service == service && session.Checkout == checkout && session.Status.IsAlive() {
			foundSession = session
		}
	}
	return foundSession
}

func (sessionHandler *SessionHandler) startAcceptingNewSessionRequests() {
	go func() {
		for {
			// I'm trying to build my session.
			// Wait here until someone requests some work
			sessionBuildRequest := <-sessionHandler.sessionRequestChan

			sessionBuildResult := sessionHandler.buildSession(sessionBuildRequest)

			sessionHandler.sessionResponseChan <- sessionBuildResult

		}
	}()
}

func (sessionHandler *SessionHandler) startAcceptingSessionCleanRequests() {
	go func() {
		for {
			sessionToClean := <-sessionHandler.sessionCleanChan
			sessionToClean.session.LogInfo("Cleaning up session")
			sessionToCleanIndex := -1
			for i, session := range sessionHandler.sessions {
				if session == sessionToClean.session {
					sessionToCleanIndex = i
				}
			}
			if sessionToCleanIndex == -1 { // No session found
				log.Fatalf("[SESSION:%s] Requested session cleanup, but not found", sessionToClean.session.UUID)
			} else {
				// sessionHandler.sessions = append(
				// 	sessionHandler.sessions[:sessionToCleanIndex],
				// 	sessionHandler.sessions[sessionToCleanIndex+1:]...,
				// )
				sessionHandler.sessions[sessionToCleanIndex].Status = sessionToClean.status
				log.Warnf("[SESSION:%s] Session cleaned up.", sessionToClean.session.UUID)
			}
		}
	}()
}

func (sessionHandler *SessionHandler) CleanupSession(session *models.Session, status models.SessionStatus) {
	sessionHandler.sessionCleanChan <- &SessionClean{
		session: session,
		status:  status,
	}
}

func (sessionHandler *SessionHandler) MarkSessionAsStarted(session *models.Session) {
	session.Status = models.SessionStatusStarted

	sessionHandler.StartSessionInactivityTimer(session)
}

func (sessionHandler *SessionHandler) MarkSessionAsBeingRequested(session *models.Session) {
	// Refreshes the inactiveAt field every time someone makes a request to this session
	session.InactiveAt = time.Now().Add(time.Second * time.Duration(session.Service.Recycle.InactivityTimeout))
	session.MaxAge = session.Service.Recycle.InactivityTimeout
}

func (sessionHandler *SessionHandler) StartSessionInactivityTimer(session *models.Session) {
	session.InactiveAt = time.Now().Add(time.Second * time.Duration(session.Service.Recycle.InactivityTimeout))
	go func() {
		for {
			if session.Status != models.SessionStatusStarted {
				return
			}

			if time.Now().After(session.InactiveAt) {
				sessionHandler.DestroySession(session)
				return
			}
			session.MaxAge--
			time.Sleep(1 * time.Second)
		}
	}()
}
