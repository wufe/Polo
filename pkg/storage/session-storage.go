package storage

import (
	"encoding/json"
	"fmt"

	"github.com/dgraph-io/badger/v3"
	"github.com/wufe/polo/pkg/logging"
	"github.com/wufe/polo/pkg/models"
	"github.com/wufe/polo/pkg/utils"
)

// Session is the session storage.
// Contains methods to access and store sessions into the database
type Session struct {
	utils.RWLocker
	database           Database
	sessions           []*models.Session
	sessionsByCategory *SessionsByCategory
	log                logging.Logger
}

// NewSession creates new database storage
func NewSession(db Database, mutexBuilder utils.MutexBuilder, logger logging.Logger) *Session {
	session := &Session{
		RWLocker:           mutexBuilder(),
		database:           db,
		sessions:           make([]*models.Session, 0),
		sessionsByCategory: newSessionsByCategory(mutexBuilder),
		log:                logger,
	}
	return session
}

// LoadSessions given an application, restores its sessions
// retrieving them from the database
func (s *Session) LoadSessions(application *Application, sessionBuilder *models.SessionBuilder) {
	sessions := []*models.Session{}
	err := s.database.GetDB().View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		prefix := []byte("session/")
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			err := item.Value(func(v []byte) error {
				var session models.Session
				err := json.Unmarshal(v, &session)
				if err != nil {
					return err
				}
				if session.Status.IsAlive() {
					sessions = append(sessions, sessionBuilder.Build(&session))
				}
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	for _, session := range sessions {
		app := application.Get(session.ApplicationName)
		session.Application = app
		session.InitializeConfiguration()
		session.InitializeL4Forwards()
		if session.Application == nil {
			s.log.Errorf("Session with id %s and application name %s could not be attached to any configured application. Shutdown it manually.", session.UUID, session.ApplicationName)
			s.Delete(session)
		} else {
			s.Lock()
			s.sessions = append(s.sessions, session)
			s.Unlock()
		}
	}
	s.log.Infof("Loaded %d sessions", len(s.sessions))
	if err != nil {
		s.log.Errorf("Error while loading sessions: %s", err.Error())
	}
}

// Add stores a session
// Database-wise works as an upsert
func (s *Session) Add(session *models.Session) {
	s.log.Tracef("Storing session %s", session.UUID)
	existing := s.GetByUUID(session.UUID)
	if existing != nil {
		return
	}
	s.Lock()
	defer s.Unlock()
	s.sessions = append(s.sessions, session)
	s.internalUpdate(session)
}

// Update updates a session.
// Database-wise works as an upsert
func (s *Session) Update(session *models.Session) {
	s.log.Tracef("Updating session %s", session.UUID)
	s.internalUpdate(session)
}

func (s *Session) internalUpdate(session *models.Session) {
	err := s.database.GetDB().Update(func(txn *badger.Txn) error {
		session.RLock()
		result, err := json.Marshal(session)
		defer session.RUnlock()
		if err != nil {
			return err
		}
		err = txn.Set([]byte(fmt.Sprintf("session/%s", session.UUID)), result)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		s.log.Errorf("Error while persisting session %s: %s", session.UUID, err.Error())
	}
}

// Delete removes a session
func (s *Session) Delete(session *models.Session) {
	s.log.Tracef("Deleting session %s", session.UUID)
	err := s.database.GetDB().Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(fmt.Sprintf("session/%s", session.UUID)))
	})
	if err != nil {
		s.log.Errorf("Error while deleting session %s: %s", session.UUID, err.Error())
	}
}

// AliveByApplicationCount retrieves the number of sessions of an application
func (s *Session) AliveByApplicationCount(application *models.Application) int {
	s.log.Trace("Getting alive sessions count by application")
	count := 0
	for _, session := range s.sessions {
		if session.Application == application && session.Status.IsAlive() {
			count++
		}
	}
	return count
}

// GetByApplicationName retrieves a slice of sessions given their app name
func (s *Session) GetByApplicationName(app string) []*models.Session {
	ret := []*models.Session{}
	s.RLock()
	sessions := s.sessions
	s.RUnlock()
	for _, session := range sessions {
		if session.ApplicationName == app {
			ret = append(ret, session)
		}
	}
	return ret
}

// GetByUUID retrieves a session given its UUID
func (s *Session) GetByUUID(uuid string) *models.Session {
	var foundSession *models.Session
	s.RLock()
	sessions := s.sessions
	s.RUnlock()
	for _, session := range sessions {
		if session.UUID == uuid {
			foundSession = session
		}
	}
	return foundSession
}

// GetAll retrieves all sessions
func (s *Session) GetAll() []*models.Session {
	s.log.Trace("Getting all alive sessions")
	s.RLock()
	filteredSessions := make([]*models.Session, 0, len(s.sessions))
	sessions := s.sessions
	s.RUnlock()
	filteredSessions = append(filteredSessions, sessions...)
	return filteredSessions
}

// GetAllAliveSessions retrieves a slice of sessions whose status is "alive".
// A session is "alive" if it can or is about to ready for being used
func (s *Session) GetAllAliveSessions() []*models.Session {
	s.log.Trace("Getting all alive sessions")
	filteredSessions := []*models.Session{}
	s.RLock()
	sessions := s.sessions
	s.RUnlock()
	for _, session := range sessions {
		status := session.GetStatus()
		if status.IsAlive() {
			filteredSessions = append(filteredSessions, session)
		}
	}
	return filteredSessions
}

// GetAllAliveApplicationSessions retrieves a slice of sessions whose status is "alive".
// A session is "alive" if it can or is about to ready for being used
// The sessions are further filtered by their application ID
func (s *Session) GetAllAliveApplicationSessions(appID string) []*models.Session {
	allSessions := s.GetAllAliveSessions()
	filteredSessions := []*models.Session{}
	for _, session := range allSessions {
		conf := session.GetConfiguration()
		id := conf.ID
		if id == appID {
			filteredSessions = append(filteredSessions, session)
		}
	}
	return filteredSessions
}

// GetAliveApplicationSessionByCheckout retrieves a single session identified by its
// status (which must be "alive") and by its checkout
func (s *Session) GetAliveApplicationSessionByCheckout(checkout string, application *models.Application) *models.Session {
	s.log.Trace("Getting alive session by checkout")
	var foundSession *models.Session
	s.RLock()
	sessions := s.sessions
	s.RUnlock()
	for _, session := range sessions {
		if session.Application == application && session.Checkout == checkout && session.Status.IsAlive() {
			foundSession = session
		}
	}
	return foundSession
}

// GetAliveApplicationSessionByCommitID retrieves a single session identified by its
// status (which must be "alive") and by its commitID
func (s *Session) GetAliveApplicationSessionByCommitID(commitID string, application *models.Application) *models.Session {
	s.log.Trace("Getting alive session by commitID")
	var foundSession *models.Session
	s.RLock()
	sessions := s.sessions
	s.RUnlock()
	for _, session := range sessions {
		if session.Application == application && session.CommitID == commitID && session.Status.IsAlive() {
			foundSession = session
		}
	}
	return foundSession
}

// GetAliveApplicationSession retrieves all "alive" sessions of an application
func (s *Session) GetAliveApplicationSession(application *models.Application) []*models.Session {
	s.log.Trace("Getting alive session in application")
	foundSessions := []*models.Session{}
	s.RLock()
	sessions := s.sessions
	s.RUnlock()
	for _, session := range sessions {
		if session.Application == application && session.Status.IsAlive() {
			foundSessions = append(foundSessions, session)
		}
	}
	return foundSessions
}

const (
	SessionCategoryFailedToStart SessionCategory = "failed_to_start"
	// Failed sessions that are acknowledged by the user
	// A session goes from SessionCategoryFailedToStart to SessionCategoryFailedToStartAcknowledged
	// when a user sees its logs from the frontend
	SessionCategoryFailedToStartAcknowledged SessionCategory = "failed_to_start_ack"
)

type SessionCategory string

type SessionsByCategory struct {
	utils.RWLocker

	Data map[SessionCategory][]*models.Session
}

func newSessionsByCategory(mutexBuilder utils.MutexBuilder) *SessionsByCategory {
	return &SessionsByCategory{
		RWLocker: mutexBuilder(),
		Data:     make(map[SessionCategory][]*models.Session),
	}
}

func (s *Session) AddSessionToCategory(category SessionCategory, session *models.Session) {
	s.sessionsByCategory.Lock()
	defer s.sessionsByCategory.Unlock()
	s.sessionsByCategory.Data[category] = append(s.sessionsByCategory.Data[category], session)
}
func (s *Session) RemoveSessionFromCategory(category SessionCategory, sessionUUID string) {
	s.sessionsByCategory.Lock()
	defer s.sessionsByCategory.Unlock()
	newSessions := []*models.Session{}
	for _, session := range s.sessionsByCategory.Data[category] {
		if session.UUID != sessionUUID {
			newSessions = append(newSessions, session)
		}
	}
	s.sessionsByCategory.Data[category] = newSessions
}

func (s *Session) GetSessionsByCategory(category SessionCategory) []*models.Session {
	s.sessionsByCategory.RLock()
	defer s.sessionsByCategory.RUnlock()

	// Find all sessions by category
	var foundSessions []*models.Session
	if sessions, exists := s.sessionsByCategory.Data[category]; exists {
		foundSessions = sessions
	} else {
		foundSessions = []*models.Session{}
	}
	return foundSessions
}

func (s *Session) GetAllSessionsNames() []string {
	s.RLock()
	defer s.RUnlock()
	names := make([]string, 0, len(s.sessions))
	for _, session := range s.sessions {
		names = append(names, session.Alias)
	}
	return names
}
