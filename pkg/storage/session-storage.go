package storage

import (
	"encoding/json"
	"fmt"

	"github.com/dgraph-io/badger/v3"
	log "github.com/sirupsen/logrus"
	"github.com/wufe/polo/pkg/models"
	"github.com/wufe/polo/pkg/utils"
)

// Session is the session storage.
// Contains methods to access and store sessions into the database
type Session struct {
	utils.RWLocker
	database Database
	sessions []*models.Session
}

// NewSession creates new database storage
func NewSession(db Database, environment utils.Environment) *Session {
	session := &Session{
		RWLocker: utils.GetMutex(environment),
		database: db,
		sessions: make([]*models.Session, 0),
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
		if session.Application == nil {
			log.Errorf("Session with id %s and application name %s could not be attached to any configured application. Shutdown it manually.", session.UUID, session.ApplicationName)
			s.Delete(session)
		} else {
			s.Lock()
			s.sessions = append(s.sessions, session)
			s.Unlock()
		}
	}
	log.Infof("Loaded %d sessions", len(s.sessions))
	if err != nil {
		log.Errorf("Error while loading sessions: %s", err.Error())
	}
}

// Add stores a session
// Database-wise works as an upsert
func (s *Session) Add(session *models.Session) {
	s.Lock()
	defer s.Unlock()
	log.Tracef("Storing session %s", session.UUID)
	s.sessions = append(s.sessions, session)
	s.internalUpdate(session)
}

// Update updates a session.
// Database-wise works as an upsert
func (s *Session) Update(session *models.Session) {
	log.Tracef("Updating session %s", session.UUID)
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
		log.Errorf("Error while persisting session %s: %s", session.UUID, err.Error())
	}
}

// Delete removes a session
func (s *Session) Delete(session *models.Session) {
	log.Tracef("Deleting session %s", session.UUID)
	err := s.database.GetDB().Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(fmt.Sprintf("session/%s", session.UUID)))
	})
	if err != nil {
		log.Errorf("Error while deleting session %s: %s", session.UUID, err.Error())
	}
}

// AliveByApplicationCount retrieves the number of sessions of an application
func (s *Session) AliveByApplicationCount(application *models.Application) int {
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

// GetAllAliveSessions retrieves a slice of sessions whose status is "alive".
// A session is "alive" if it can or is about to ready for being used
func (s *Session) GetAllAliveSessions() []*models.Session {
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

// GetAliveApplicationSessionByCheckout retrieves a single session identified by its
// status (which must be "alvie") and by its checkout
func (s *Session) GetAliveApplicationSessionByCheckout(checkout string, application *models.Application) *models.Session {
	var foundSession *models.Session
	s.RLock()
	sessions := s.sessions
	s.RUnlock()
	for _, session := range sessions {
		if session.Application == application && session.CommitID == checkout && session.Status.IsAlive() {
			foundSession = session
		}
	}
	return foundSession
}
