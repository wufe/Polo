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
	if buildResult.Result == SessionBuildResultSucceeded {
		sessionHandler.sessions = append(sessionHandler.sessions, buildResult.Session)
	}
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

			// Check if someone else just requested the same type of session
			// looking through all open session and comparing services and checkouts
			sessionAlreadyBeingBuilt := sessionHandler.GetAliveServiceSessionByCheckout(
				sessionBuildRequest.Checkout,
				sessionBuildRequest.Service,
			)

			var result *SessionBuildResult

			if sessionAlreadyBeingBuilt != nil {
				log.Infof(
					"Another session with the UUID %s has already being requested for checkout %s",
					sessionAlreadyBeingBuilt.UUID,
					sessionBuildRequest.Checkout,
				)
				result = &SessionBuildResult{
					Result:  SessionBuildResultSucceeded,
					Session: sessionAlreadyBeingBuilt,
				}
			} else {
				sessionBuildResult := sessionHandler.buildSession(sessionBuildRequest)
				// Oke, session has been created; Or Nope, it failed miserably
				result = sessionBuildResult
			}

			sessionHandler.sessionResponseChan <- result

		}
	}()
}

func (sessionHandler *SessionHandler) startAcceptingSessionCleanRequests() {
	go func() {
		for {
			sessionToClean := <-sessionHandler.sessionCleanChan
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

// TODO: Call this on net requests
func (sessionHandler *SessionHandler) MarkSessionAsBeingRequested(session *models.Session) {
	// Refreshes the inactiveAt field every time someone makes a request to this session
	session.InactiveAt = time.Now().Add(time.Second * time.Duration(session.Service.Recycle.InactivityTimeout))
}

func (sessionHandler *SessionHandler) StartSessionInactivityTimer(session *models.Session) {
	session.InactiveAt = time.Now().Add(time.Second * time.Duration(session.Service.Recycle.InactivityTimeout))
	go func() {
		for {
			if time.Now().After(session.InactiveAt) {
				sessionHandler.DestroySession(session)
				return
			}
			time.Sleep(1 * time.Second)
		}
	}()
}
