package storage

import "github.com/wufe/polo/models"

type Session struct {
	sessions []*models.Session
}

func NewSession() *Session {
	return &Session{
		sessions: make([]*models.Session, 0),
	}
}

func (s *Session) Add(sessions ...*models.Session) {
	s.sessions = append(s.sessions, sessions...)
}

func (s *Session) AliveByApplicationCount(application *models.Application) int {
	count := 0
	for _, session := range s.sessions {
		if session.Application == application && session.Status.IsAlive() {
			count++
		}
	}
	return count
}

func (s *Session) GetByUUID(uuid string) *models.Session {
	var foundSession *models.Session
	for _, session := range s.sessions {
		if session.UUID == uuid {
			foundSession = session
		}
	}
	return foundSession
}

func (s *Session) GetAllAliveSessions() []*models.Session {
	filteredSessions := []*models.Session{}
	for _, session := range s.sessions {
		if session.Status.IsAlive() {
			filteredSessions = append(filteredSessions, session)
		}
	}
	return filteredSessions
}

func (s *Session) GetAliveApplicationSessionByCheckout(checkout string, application *models.Application) *models.Session {
	var foundSession *models.Session
	for _, session := range s.sessions {
		if session.Application == application && session.CommitID == checkout && session.Status.IsAlive() {
			foundSession = session
		}
	}
	return foundSession
}
