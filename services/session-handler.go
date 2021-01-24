package services

import (
	log "github.com/sirupsen/logrus"
	"github.com/wufe/polo/models"
)

type SessionHandler struct {
	configuration       *models.RootConfiguration
	sessions            []*models.Session
	sessionRequestChan  chan *SessionBuildInput
	sessionResponseChan chan *SessionBuildResult
	sessionCleanChan    chan *models.Session
}

func NewSessionHandler(configuration *models.RootConfiguration) *SessionHandler {
	sessionHandler := &SessionHandler{
		configuration:       configuration,
		sessions:            []*models.Session{},
		sessionRequestChan:  make(chan *SessionBuildInput),
		sessionResponseChan: make(chan *SessionBuildResult),
		sessionCleanChan:    make(chan *models.Session),
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
		if session.Status != models.SessionStatusStartFailed {
			filteredSessions = append(filteredSessions, session)
		}
	}
	return filteredSessions
}

func (sessionHandler *SessionHandler) GetServiceSessionByCheckout(checkout string, service *models.Service) *models.Session {
	var foundSession *models.Session
	for _, session := range sessionHandler.sessions {
		if session.Service == service && session.Checkout == checkout {
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
			sessionAlreadyBeingBuilt := sessionHandler.GetServiceSessionByCheckout(
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
				if session == sessionToClean {
					sessionToCleanIndex = i
				}
			}
			if sessionToCleanIndex == -1 { // No session found
				log.Fatalf("[SESSION:%s] Requested session cleanup, but not found", sessionToClean.UUID)
			} else {
				// sessionHandler.sessions = append(
				// 	sessionHandler.sessions[:sessionToCleanIndex],
				// 	sessionHandler.sessions[sessionToCleanIndex+1:]...,
				// )
				sessionHandler.sessions[sessionToCleanIndex].Status = models.SessionStatusStartFailed
				log.Warnf("[SESSION:%s] Session cleaned up.", sessionToClean.UUID)
			}
		}
	}()
}

func (sessionHandler *SessionHandler) CleanupSession(session *models.Session) {
	sessionHandler.sessionCleanChan <- session
}
