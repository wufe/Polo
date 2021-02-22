package models

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/go-git/go-git/v5/plumbing/object"
	log "github.com/sirupsen/logrus"
	"github.com/wufe/polo/pkg/utils"
)

const (
	SessionStatusStarting    SessionStatus = "starting"
	SessionStatusStarted     SessionStatus = "started"
	SessionStatusStartFailed SessionStatus = "start_failed"
	SessionStatusStopFailed  SessionStatus = "stop_failed"
	SessionStatusStopping    SessionStatus = "stopping"
	SessionStatusStopped     SessionStatus = "stopped"
	SessionStatusDegraded    SessionStatus = "degraded"

	LogTypeStdin  LogType = "stdin"
	LogTypeStdout LogType = "stdout"
	LogTypeStderr LogType = "stderr"

	KillReasonNone              KillReason = ""
	KillReasonStopped           KillReason = "stopped"
	KillReasonBuildFailed       KillReason = "build_failed"
	KillReasonHealthcheckFailed KillReason = "healthcheck_failed"

	SessionBuildContextKey string = "build"
)

type SessionStatus string

func (status SessionStatus) IsAlive() bool {
	return status != SessionStatusStartFailed &&
		status != SessionStatusStopFailed &&
		status != SessionStatusStopped &&
		status != SessionStatusStopping
}

type Session struct {
	utils.RWLocker  `json:"-"`
	UUID            string       `json:"uuid"`
	ShortUUID       string       `json:"-"`
	Name            string       `json:"name"`
	Target          string       `json:"target"`
	Port            int          `json:"port"`
	ApplicationName string       `json:"applicationName"`
	Application     *Application `json:"-"`
	configuration   ApplicationConfiguration
	Status          SessionStatus `json:"status"`
	Logs            []Log         `json:"-"`
	CommitID        string        `json:"commitID"` // The object to be checked out (branch/tag/commit id)
	Checkout        string        `json:"checkout"`
	Commit          object.Commit `json:"commit"`
	MaxAge          int           `json:"maxAge"`
	InactiveAt      time.Time     `json:"-"`
	Folder          string        `json:"folder"`
	Variables       Variables     `json:"variables"`
	Metrics         []Metric      `json:"metrics"`
	startupRetries  int
	killReason      KillReason    `json:"-"`
	Context         *contextStore `json:"-"`
}

type KillReason string

type Variables map[string]string

func (v Variables) ApplyTo(str string) string {
	for key, value := range v {
		str = strings.ReplaceAll(str, fmt.Sprintf("{{%s}}", key), value)
	}
	return str
}

func NewSession(
	session *Session,
) *Session {
	session.ShortUUID = strings.Split(session.UUID, "-")[0]
	session.RWLocker = utils.GetMutex()
	if session.ApplicationName == "" {
		session.ApplicationName = session.Application.GetConfiguration().Name
	}
	session.Status = SessionStatusStarting
	if session.Logs == nil {
		session.Logs = []Log{}
	}
	if len(session.Variables) == 0 {
		session.Variables = make(map[string]string)
	}
	if session.Metrics == nil {
		session.Metrics = []Metric{}
	}
	session.killReason = KillReasonNone
	session.Context = NewContextStore()
	if session.Application != nil {
		session.configuration = session.getMatchingConfiguration()
	}
	return session
}

func (session *Session) GetConfiguration() ApplicationConfiguration {
	session.RLock()
	defer session.RUnlock()
	return session.configuration
}

func (session *Session) SetConfiguration(conf ApplicationConfiguration) {
	session.Lock()
	defer session.Unlock()
	session.Application.SetConfiguration(conf)
	session.configuration = session.getMatchingConfiguration()
}

func (session *Session) getMatchingConfiguration() ApplicationConfiguration {
	branches := session.Application.configuration.Branches
	baseConfig := session.Application.GetConfiguration()
	if branches == nil {
		return baseConfig
	}
	checkout := session.Checkout
	found := false
	var matchingConf BranchConfiguration
	for _, conf := range branches {
		testRE := regexp.MustCompile(conf.Test)
		if testRE.MatchString(checkout) {
			matchingConf = conf.BranchConfiguration
			found = true
			break
		}
	}
	if !found {
		return baseConfig
	}
	baseConfig.OverrideWith(matchingConf)
	return baseConfig
}

func (session *Session) LogCritical(message string) {
	session.Lock()
	log.Errorf(fmt.Sprintf("\t[%s]: %s", session.ShortUUID, message))
	defer session.Unlock()
	session.Logs = append(
		session.Logs,
		NewLog(message, LogTypeCritical),
	)
}

func (session *Session) LogError(message string) {
	session.Lock()
	log.Errorf(fmt.Sprintf("\t[%s]: %s", session.ShortUUID, message))
	defer session.Unlock()
	session.Logs = append(
		session.Logs,
		NewLog(message, LogTypeError),
	)
}

func (session *Session) LogWarn(message string) {
	session.Lock()
	log.Warnf(fmt.Sprintf("\t[%s]: %s", session.ShortUUID, message))
	defer session.Unlock()
	session.Logs = append(
		session.Logs,
		NewLog(message, LogTypeWarn),
	)
}

func (session *Session) LogInfo(message string) {
	session.Lock()
	log.Infof(fmt.Sprintf("\t[%s]: %s", session.ShortUUID, message))
	defer session.Unlock()
	session.Logs = append(
		session.Logs,
		NewLog(message, LogTypeInfo),
	)
}

func (session *Session) LogDebug(message string) {
	session.Lock()
	log.Debugf(fmt.Sprintf("\t[%s]: %s", session.ShortUUID, message))
	defer session.Unlock()
	session.Logs = append(
		session.Logs,
		NewLog(message, LogTypeDebug),
	)
}

func (session *Session) LogTrace(message string) {
	session.Lock()
	log.Tracef(fmt.Sprintf("\t[%s]: %s", session.ShortUUID, message))
	defer session.Unlock()
	session.Logs = append(
		session.Logs,
		NewLog(message, LogTypeTrace),
	)
}

func (session *Session) LogStdin(message string) {
	session.Lock()
	log.Infof(fmt.Sprintf("\t\t[%s (stdin)>]: %s", session.ShortUUID, message))
	defer session.Unlock()
	session.Logs = append(
		session.Logs,
		NewLog(message, LogTypeStdin),
	)
}

func (session *Session) LogStdout(message string) {
	session.Lock()
	log.Infof(fmt.Sprintf("\t\t[%s (stdout)>]: %s", session.ShortUUID, message))
	defer session.Unlock()
	session.Logs = append(
		session.Logs,
		NewLog(message, LogTypeStdout),
	)
}

func (session *Session) LogStderr(message string) {
	session.Lock()
	log.Infof(fmt.Sprintf("\t\t[%s (stderr)>]: %s", session.ShortUUID, message))
	defer session.Unlock()
	session.Logs = append(
		session.Logs,
		NewLog(message, LogTypeStderr),
	)
}

func (session *Session) MarkAsBeingRequested() {
	conf := session.GetConfiguration()
	if session.GetMaxAge() > -1 {
		// Refreshes the inactiveAt field every time someone makes a request to this session
		session.SetInactiveAt(time.Now().Add(time.Second * time.Duration(conf.Recycle.InactivityTimeout)))
		session.SetMaxAge(conf.Recycle.InactivityTimeout)
	}
}

func (session *Session) SetStatus(status SessionStatus) {
	session.Lock()
	defer session.Unlock()
	session.Status = status
}

func (session *Session) GetStatus() SessionStatus {
	session.Lock()
	defer session.Unlock()
	return session.Status
}

func (session *Session) DecreaseMaxAge() {
	session.Lock()
	defer session.Unlock()
	session.MaxAge--
}

func (session *Session) GetMaxAge() int {
	session.Lock()
	defer session.Unlock()
	return session.MaxAge
}

func (session *Session) SetMaxAge(age int) {
	session.Lock()
	defer session.Unlock()
	session.MaxAge = age
}

func (session *Session) GetInactiveAt() time.Time {
	session.Lock()
	defer session.Unlock()
	return session.InactiveAt
}

func (session *Session) SetInactiveAt(at time.Time) {
	session.Lock()
	defer session.Unlock()
	session.InactiveAt = at
}

func (session *Session) GetStartupRetriesCount() int {
	session.Lock()
	defer session.Unlock()
	return session.startupRetries
}

func (session *Session) SetStartupRetriesCount(retries int) {
	session.Lock()
	defer session.Unlock()
	session.startupRetries = retries
}

func (session *Session) IncStartupRetriesCount() {
	session.Lock()
	defer session.Unlock()
	session.startupRetries++
}

func (session *Session) ResetStartupRetriesCount() {
	session.Lock()
	defer session.Unlock()
	session.startupRetries = 0
}

func (session *Session) GetKillReason() KillReason {
	session.Lock()
	defer session.Unlock()
	return session.killReason
}

func (session *Session) SetKillReason(reason KillReason) {
	session.Lock()
	defer session.Unlock()
	session.killReason = reason
}

func (session *Session) SetVariable(k string, v string) {
	session.Lock()
	defer session.Unlock()
	session.Variables[k] = v
}

func (session *Session) ResetVariables() {
	session.Lock()
	defer session.Unlock()
	session.Variables = make(map[string]string)
}
