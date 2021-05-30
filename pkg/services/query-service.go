package services

import (
	"fmt"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/object"
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

// GetMatchingCheckoutBySmartUrl
// The rawInput parameter is without prefix "/s/"
func (s *QueryService) GetMatchingCheckoutBySmartUrl(rawInput string) (checkout string, application string, path string, found bool) {
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

func (s *QueryService) GetMatchingCheckoutByPermalink(rawInput string) (checkout string, application string, path string, found bool) {
	// Format: <app-hash>/<commit-id>/<path>?
	if !strings.Contains(rawInput, "/") {
		return "", "", "", false
	}
	var foundApp *models.Application
	apps := s.applicationStorage.GetAll()
	appHashPrefix := ""
	for _, app := range apps {
		conf := app.GetConfiguration()
		appHashPrefix = conf.Hash + "/"
		if strings.HasPrefix(rawInput, appHashPrefix) {
			rawInput = strings.TrimPrefix(rawInput, appHashPrefix)
			foundApp = app
			application = conf.Name
			break
		}
	}
	if foundApp == nil {
		return "", "", "", false
	}
	// "rawInput" is now stripped of the app-hash prefix
	// New format: <commit-id>/<path>?
	// where <path> may contain other "/"s
	// That's why we are using SplitN with N = 2
	// e.g. <commit-id>/1/2/3 -> [<commit-id>, "1/2/3"]
	chunks := strings.SplitN(rawInput, "/", 2)
	commitID := chunks[0]
	var commitMap map[string]*object.Commit
	foundApp.WithRLock(func(a *models.Application) {
		commitMap = a.CommitMap
	})
	_, found = commitMap[commitID]
	if !found {
		// Does not exist a commit with that ID
		return "", "", "", false
	}
	checkout = commitID
	if len(chunks) > 1 {
		path = chunks[1]
	}
	return checkout, application, path, found
}

func (s *QueryService) GetFailedSessions() []*models.Session {
	return s.sessionStorage.GetSessionsByCategory(storage.SessionCategoryFailedToStart)
}

func (s *QueryService) GetSeenFailedSessions() []*models.Session {
	return s.sessionStorage.GetSessionsByCategory(storage.SessionCategoryFailedToStartAcknowledged)
}

// Retrieve a failed session, search through seen and unseen
func (s *QueryService) GetFailedSession(uuid string) (*models.Session, error) {
	unacknowledged := s.GetFailedSessions()
	var foundSession *models.Session
	for _, session := range unacknowledged {
		if session.UUID == uuid {
			foundSession = session
		}
	}
	if foundSession == nil {
		acknowledged := s.GetSeenFailedSessions()
		for _, session := range acknowledged {
			if session.UUID == uuid {
				foundSession = session
			}
		}
		if foundSession == nil {
			return nil, ErrSessionNotFound
		}
	}
	return foundSession, nil
}

func (s *QueryService) GetFailedSessionLogs(uuid string) ([]models.Log, error) {
	session, err := s.GetFailedSession(uuid)
	if err != nil {
		return nil, ErrSessionNotFound
	}
	return session.GetLogs(), nil
}

func (s *QueryService) MarkFailedSessionAsSeen(uuid string) {
	session, err := s.GetFailedSession(uuid)
	if err == nil {
		s.sessionStorage.RemoveSessionFromCategory(storage.SessionCategoryFailedToStart, uuid)
		s.sessionStorage.AddSessionToCategory(storage.SessionCategoryFailedToStartAcknowledged, session)
	}
}
