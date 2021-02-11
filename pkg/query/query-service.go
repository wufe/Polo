package query

import (
	"errors"

	"github.com/wufe/polo/pkg/models"
	"github.com/wufe/polo/pkg/storage"
)

type Service struct {
	isDev              bool
	sessionStorage     *storage.Session
	applicationStorage *storage.Application
}

var (
	ErrSessionNotFound error = errors.New("Session not found")
)

func NewService(isDev bool, storage *storage.Session, applicationStorage *storage.Application) *Service {
	s := &Service{
		isDev:              isDev,
		sessionStorage:     storage,
		applicationStorage: applicationStorage,
	}
	return s
}

func (s *Service) GetAllApplications() []*models.Application {
	return s.applicationStorage.GetAll()
}

func (s *Service) GetAllAliveSessions() []*models.Session {
	return s.sessionStorage.GetAllAliveSessions()
}

func (s *Service) GetSession(uuid string) *models.Session {
	var foundSession *models.Session
	for _, session := range s.sessionStorage.GetAllAliveSessions() {
		if session.UUID == uuid {
			foundSession = session
		}
	}

	return foundSession
}

func (s *Service) GetSessionAge(uuid string) (int, error) {
	session := s.GetSession(uuid)
	if session == nil {
		return 0, ErrSessionNotFound
	}
	return session.GetMaxAge(), nil
}

func (s *Service) GetSessionMetrics(uuid string) ([]*models.Metric, error) {
	session := s.GetSession(uuid)
	if session == nil {
		return nil, ErrSessionNotFound
	}
	return session.Metrics, nil
}

func (s *Service) GetSessionLogsAndStatus(uuid string, lastLogUUID string) ([]models.Log, models.SessionStatus, error) {
	session := s.GetSession(uuid)
	if session == nil {
		return nil, models.SessionStatusStarting, ErrSessionNotFound
	}

	logs := session.Logs
	if lastLogUUID != "" && lastLogUUID != "<none>" {
		logs = []models.Log{}
		afterLastLog := false
		for _, log := range session.Logs {
			if afterLastLog {
				logs = append(logs, log)
			}
			if log.UUID == lastLogUUID {
				afterLastLog = true
			}
		}
	}

	return logs, session.Status, nil
}
