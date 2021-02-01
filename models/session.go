package models

import (
	"time"
)

const (
	SessionStatusStarting    SessionStatus = "starting"
	SessionStatusStarted     SessionStatus = "started"
	SessionStatusStartFailed SessionStatus = "start_failed"
	SessionStatusStopFailed  SessionStatus = "stop_failed"
	SessionStatusStopping    SessionStatus = "stopping"
	SessionStatusStopped     SessionStatus = "stopped"

	LogTypeStdin  LogType = "stdin"
	LogTypeStdout LogType = "stdout"
	LogTypeStderr LogType = "stderr"
)

type SessionStatus string

func (status SessionStatus) IsAlive() bool {
	return status != SessionStatusStartFailed &&
		status != SessionStatusStopFailed &&
		status != SessionStatusStopped
}

type Session struct {
	UUID         string            `json:"uuid"`
	Name         string            `json:"name"`
	Target       string            `json:"target"`
	Port         int               `json:"port"`
	Service      *Service          `json:"service"`
	Status       SessionStatus     `json:"status"`
	Logs         []Log             `json:"logs"`
	CommitID     string            `json:"commitID"` // The object to be checked out (branch/tag/commit id)
	Checkout     string            `json:"checkout"`
	Done         chan struct{}     `json:"-"`
	MaxAge       int               `json:"maxAge"`
	InactiveAt   time.Time         `json:"-"`
	Folder       string            `json:"folder"`
	CommandsLogs []string          `json:"commandsLogs"`
	Variables    map[string]string `json:"variables"`
}

func NewSession(
	session *Session,
) *Session {
	session.CommandsLogs = []string{}
	session.Variables = make(map[string]string)
	return session
}

func (session *Session) LogCritical(message string) {
	session.Logs = append(session.Logs, Log{
		Message: message,
		Type:    LogTypeCritical,
	})
}

func (session *Session) LogError(message string) {
	session.Logs = append(session.Logs, Log{
		Message: message,
		Type:    LogTypeError,
	})
}

func (session *Session) LogWarn(message string) {
	session.Logs = append(session.Logs, Log{
		Message: message,
		Type:    LogTypeWarn,
	})
}

func (session *Session) LogInfo(message string) {
	session.Logs = append(session.Logs, Log{
		Message: message,
		Type:    LogTypeInfo,
	})
}

func (session *Session) LogDebug(message string) {
	session.Logs = append(session.Logs, Log{
		Message: message,
		Type:    LogTypeDebug,
	})
}

func (session *Session) LogTrace(message string) {
	session.Logs = append(session.Logs, Log{
		Message: message,
		Type:    LogTypeTrace,
	})
}

func (session *Session) LogStdin(message string) {
	session.Logs = append(session.Logs, Log{
		Message: message,
		Type:    LogTypeStdin,
	})
}

func (session *Session) LogStdout(message string) {
	session.Logs = append(session.Logs, Log{
		Message: message,
		Type:    LogTypeStdout,
	})
}

func (session *Session) LogStderr(message string) {
	session.Logs = append(session.Logs, Log{
		Message: message,
		Type:    LogTypeStderr,
	})
}
