package storage

import (
	"encoding/json"
	"fmt"

	"github.com/dgraph-io/badger/v3"
	log "github.com/sirupsen/logrus"
	"github.com/wufe/polo/pkg/models"
)

type Session struct {
	database *Database
	sessions []*models.Session
}

func NewSession(db *Database) *Session {
	session := &Session{
		database: db,
		sessions: make([]*models.Session, 0),
	}
	return session
}

func (s *Session) LoadSessions(application *Application) {
	sessions := []*models.Session{}
	err := s.database.DB.View(func(txn *badger.Txn) error {
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
				sessions = append(sessions, models.NewSession(&session))
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	for _, session := range sessions {
		session.Application = application.Get(session.ApplicationName)
		if session.Application == nil {
			log.Errorf("Session with id %s and application name %s could not be attached to any configured application. Shutdown it manually.", session.UUID, session.ApplicationName)
			s.Delete(session)
		} else {
			s.sessions = append(s.sessions, session)
		}
	}
	log.Infof("Loaded %d sessions", len(s.sessions))
	if err != nil {
		log.Errorf("Error while loading sessions: %s", err.Error())
	}
}

func (s *Session) Add(session *models.Session) {
	log.Tracef("Storing session %s", session.UUID)
	s.sessions = append(s.sessions, session)
	s.internalUpdate(session)
}

func (s *Session) Update(session *models.Session) {
	log.Tracef("Updating session %s", session.UUID)
	s.internalUpdate(session)
}

func (s *Session) internalUpdate(session *models.Session) {
	err := s.database.DB.Update(func(txn *badger.Txn) error {
		session.Lock()
		result, err := json.Marshal(session)
		defer session.Unlock()
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

func (s *Session) Delete(session *models.Session) {
	log.Tracef("Deleting session %s", session.UUID)
	err := s.database.DB.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(fmt.Sprintf("session/%s", session.UUID)))
	})
	if err != nil {
		log.Errorf("Error while deleting session %s: %s", session.UUID, err.Error())
	}
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
