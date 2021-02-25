package services

import (
	"strings"

	"github.com/wufe/polo/pkg/background/queues"
	"github.com/wufe/polo/pkg/models"
	"github.com/wufe/polo/pkg/storage"
)

type QueryService struct {
	isDev              bool
	sessionStorage     *storage.Session
	applicationStorage *storage.Application
}

func NewQueryService(isDev bool, storage *storage.Session, applicationStorage *storage.Application) *QueryService {
	s := &QueryService{
		isDev:              isDev,
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

func (s *QueryService) GetSession(uuid string) *models.Session {
	var foundSession *models.Session
	for _, session := range s.sessionStorage.GetAllAliveSessions() {
		if session.UUID == uuid {
			foundSession = session
		}
	}

	return foundSession
}

func (s *QueryService) GetSessionAge(uuid string) (int, error) {
	session := s.GetSession(uuid)
	if session == nil {
		return 0, ErrSessionNotFound
	}
	return session.GetMaxAge(), nil
}

func (s *QueryService) GetSessionMetrics(uuid string) ([]models.Metric, error) {
	session := s.GetSession(uuid)
	if session == nil {
		return nil, ErrSessionNotFound
	}
	return session.Metrics, nil
}

func (s *QueryService) GetSessionLogsAndStatus(uuid string, lastLogUUID string) ([]models.Log, models.SessionStatus, error) {
	session := s.GetSession(uuid)
	if session == nil {
		return nil, models.SessionStatusStarting, ErrSessionNotFound
	}

	session.Lock()
	defer session.Unlock()

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

func (s *QueryService) GetMatchingCheckout(rawInput string) (queues.SessionBuildInput, bool) {
	rawInput = strings.ToLower(rawInput)
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
		return queues.SessionBuildInput{}, false
	}
	var objectsToHashMap map[string]string
	defaultApp.WithRLock(func(a *models.Application) {
		objectsToHashMap = a.ObjectsToHashMap
	})
	for k, v := range objectsToHashMap {
		if strings.ToLower(k) == rawInput {
			return queues.SessionBuildInput{
				Application: defaultApp,
				Checkout:    v,
			}, true
		}
	}
	return queues.SessionBuildInput{}, false

}
