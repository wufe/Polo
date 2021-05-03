package services

import (
	"errors"
	"fmt"
	"strings"

	"github.com/wufe/polo/pkg/models"
	"github.com/wufe/polo/pkg/models/output"
	"github.com/wufe/polo/pkg/storage"
	"github.com/wufe/polo/pkg/utils"
)

type QueryService struct {
	isDev              bool
	sessionStorage     *storage.Session
	applicationStorage *storage.Application
}

func NewQueryService(environment utils.Environment, storage *storage.Session, applicationStorage *storage.Application) *QueryService {
	s := &QueryService{
		isDev:              environment.IsDev(),
		sessionStorage:     storage,
		applicationStorage: applicationStorage,
	}
	return s
}

func (s *QueryService) GetAllApplications() []*models.Application {
	return s.applicationStorage.GetAll()
}

func (s *QueryService) GetAllAliveSessions() []*models.Session {
	return s.sessionStorage.GetAllAliveSessions()
}

func (s *QueryService) GetAliveSession(uuid string) *models.Session {
	var foundSession *models.Session
	for _, session := range s.sessionStorage.GetAllAliveSessions() {
		if session.UUID == uuid {
			foundSession = session
		}
	}

	return foundSession
}

func (s *QueryService) GetSessionStatus(uuid string) (output.SessionStatus, error) {
	session := s.sessionStorage.GetByUUID(uuid)
	if session == nil {
		return output.SessionStatus{}, ErrSessionNotFound
	}
	return models.MapSessionStatus(session), nil
}

func (s *QueryService) GetSessionMetrics(uuid string) ([]models.Metric, error) {
	session := s.GetAliveSession(uuid)
	if session == nil {
		return nil, ErrSessionNotFound
	}
	return session.Metrics, nil
}

func (s *QueryService) GetSessionLogsAndStatus(uuid string, lastLogUUID string) ([]models.Log, models.SessionStatus, error) {
	session := s.GetAliveSession(uuid)
	if session == nil {
		return nil, models.SessionStatusStarting, ErrSessionNotFound
	}

	sessionLogs := session.GetLogs()
	retLogs := sessionLogs
	if lastLogUUID != "" && lastLogUUID != "<none>" {
		retLogs = []models.Log{}
		afterLastLog := false
		for _, log := range sessionLogs {
			if afterLastLog {
				retLogs = append(retLogs, log)
			}
			if log.UUID == lastLogUUID {
				afterLastLog = true
			}
		}
	}

	return retLogs, session.Status, nil
}

// GetMatchingCheckout
// The rawInput parameter is without prefix "/s/"
func (s *QueryService) GetMatchingCheckout(rawInput string) (checkout string, application string, path string, found bool) {
	var defaultApp *models.Application
	apps := s.applicationStorage.GetAll()
	for _, app := range apps {
		conf := app.GetConfiguration()
		if conf.IsDefault {
			defaultApp = app
			break
		}
	}
	if defaultApp == nil {
		return "", "", "", false
	}
	var objectsToHashMap map[string]string
	defaultApp.WithRLock(func(a *models.Application) {
		objectsToHashMap = a.ObjectsToHashMap
	})
	for k := range objectsToHashMap {
		if k == rawInput {
			// In case the url is formed like /s/<branch>
			return rawInput, defaultApp.GetConfiguration().Name, "", true
		} else if strings.HasPrefix(rawInput, k+"/") {
			// In case the url is formed like /s/<branch>/<path>
			path := strings.Replace(rawInput, fmt.Sprintf(`%s/`, k), "", 1)
			return k, defaultApp.GetConfiguration().Name, path, true
		}
	}
	return "", "", "", false
}

func (s *QueryService) GetFailedSessions() []*models.Session {
	return s.sessionStorage.GetSessionsByCategory(storage.SessionCategoryFailedToStart)
}

func (s *QueryService) GetFailedSessionLogs(uuid string) ([]models.Log, error) {
	failedSessions := s.sessionStorage.GetSessionsByCategory(storage.SessionCategoryFailedToStart)
	var foundSession *models.Session
	for _, session := range failedSessions {
		if session.UUID == uuid {
			foundSession = session
		}
	}
	if foundSession == nil {
		return nil, errors.New("Session not found")
	}
	return foundSession.GetLogs(), nil
}
