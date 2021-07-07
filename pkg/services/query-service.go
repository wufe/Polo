package services

import (
	"fmt"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/wufe/polo/pkg/logging"
	"github.com/wufe/polo/pkg/models"
	"github.com/wufe/polo/pkg/models/output"
	"github.com/wufe/polo/pkg/storage"
	"github.com/wufe/polo/pkg/utils"
)

type QueryService struct {
	isDev              bool
	sessionStorage     *storage.Session
	applicationStorage *storage.Application
	log                logging.Logger
}

func NewQueryService(environment utils.Environment, storage *storage.Session, applicationStorage *storage.Application, log logging.Logger) *QueryService {
	s := &QueryService{
		isDev:              environment.IsDev(),
		sessionStorage:     storage,
		applicationStorage: applicationStorage,
		log:                log,
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
func (s *QueryService) GetMatchingCheckoutBySmartUrl(rawInput string) (checkout string, application string, path string, found bool, foundSession *models.Session) {
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
		return "", "", "", false, nil
	}

	// First of all, we check for a RUNNING (started) session with the same checkout
	sessions := s.sessionStorage.GetAliveApplicationSession(defaultApp)
	for _, session := range sessions {
		if session.Status == models.SessionStatusStarted {
			if session.Checkout == rawInput {
				// In case the url is formed like /s/<branch>
				return rawInput, defaultApp.GetConfiguration().Name, "", true, session
			} else if strings.HasPrefix(rawInput, session.Checkout+"/") {
				// In case the url is formed like /s/<branch>/<path>
				path := strings.Replace(rawInput, fmt.Sprintf(`%s/`, session.Checkout), "", 1)
				return session.Checkout, defaultApp.GetConfiguration().Name, path, true, session
			}
		}
	}

	// Then we check by all existing objects (tag, branch, commit)
	// through the objectsToHashMap
	var objectsToHashMap map[string]string
	defaultApp.WithRLock(func(a *models.Application) {
		objectsToHashMap = a.ObjectsToHashMap
	})
	for k := range objectsToHashMap {
		if k == rawInput {
			// In case the url is formed like /s/<branch>
			return rawInput, defaultApp.GetConfiguration().Name, "", true, nil
		} else if strings.HasPrefix(rawInput, k+"/") {
			// In case the url is formed like /s/<branch>/<path>
			path := strings.Replace(rawInput, fmt.Sprintf(`%s/`, k), "", 1)
			return k, defaultApp.GetConfiguration().Name, path, true, nil
		}
	}
	return "", "", "", false, nil
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

func (s *QueryService) GetMatchingCheckoutByForwardLink(rawInput string) (checkout string, application string, path string, found bool) {
	// Format: <app-hash>?/(<commit-id>|<branch-name>)?/<path>?
	apps := s.applicationStorage.GetAll()
	var specificApp *models.Application
	var specificAppConfiguration models.ApplicationConfiguration
	var defaultApp *models.Application
	var defaultAppConfiguration models.ApplicationConfiguration
	for _, app := range apps {
		conf := app.GetConfiguration()
		// Store app if is default,
		// in case no specific app is found
		if conf.IsDefault {
			defaultApp = app
			defaultAppConfiguration = conf
		}
		// Try to match app against its app hash
		appHashPrefix := conf.Hash + "/"
		if strings.HasPrefix(rawInput, appHashPrefix) {
			// Remove app hash prefix from the raw input
			rawInput = strings.TrimPrefix(rawInput, appHashPrefix)
			specificApp = app
			specificAppConfiguration = conf
		}
	}
	// Populate foundApp with the corresponding app or the default app (as fallback)
	var foundApp *models.Application = specificApp
	var foundAppName string = specificAppConfiguration.Name
	var foundAppConfiguration models.ApplicationConfiguration = specificAppConfiguration
	if foundApp == nil {
		foundApp = defaultApp
		foundAppName = defaultAppConfiguration.Name
		foundAppConfiguration = defaultAppConfiguration
	}
	// If no app has been set as default, return negatives result
	if foundApp == nil {
		return "", "", "", false
	}
	// Format: (<commit-id>|<branch-name>)?/<path>?
	// <branch-name> may contain other "/"s
	// <path> may contain other "/"s
	var objectsToHashMap map[string]string
	foundApp.WithRLock(func(a *models.Application) {
		objectsToHashMap = a.ObjectsToHashMap
	})

	for commitOrName := range objectsToHashMap {
		commitOrNamePrefix := commitOrName
		if strings.HasPrefix(rawInput, commitOrNamePrefix) {
			// We found a match
			// Strip the prefix and continue
			rawInput := strings.TrimPrefix(rawInput, commitOrNamePrefix)
			return commitOrName, foundAppName, rawInput, true
		}
	}

	// Format: /<path>?
	// No commits, branchs, tags or objects in general found.
	// Try to fallback on the branch flagged as "main"
	var branchesMap map[string]*models.Branch
	foundApp.WithRLock(func(a *models.Application) {
		branchesMap = a.BranchesMap
	})
	for branchName := range branchesMap {
		if foundAppConfiguration.Branches.BranchIsMain(branchName, s.log) {
			return branchName, foundAppName, rawInput, true
		}
	}
	return "", "", "", false
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
