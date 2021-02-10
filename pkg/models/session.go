package models

import (
	"fmt"
	"strings"
	"time"
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
)

type SessionStatus string

func (status SessionStatus) IsAlive() bool {
	return status != SessionStatusStartFailed &&
		status != SessionStatusStopFailed &&
		status != SessionStatusStopped
}

type Session struct {
	UUID            string        `json:"uuid"`
	Name            string        `json:"name"`
	Target          string        `json:"target"`
	Port            int           `json:"port"`
	ApplicationName string        `json:"applicationName"`
	Application     *Application  `json:"application"`
	Status          SessionStatus `json:"status"`
	Logs            []Log         `json:"-"`
	CommitID        string        `json:"commitID"` // The object to be checked out (branch/tag/commit id)
	Checkout        string        `json:"checkout"`
	MaxAge          int           `json:"maxAge"`
	InactiveAt      time.Time     `json:"-"`
	Folder          string        `json:"folder"`
	CommandsLogs    []string      `json:"commandsLogs"`
	Variables       Variables     `json:"variables"`
}

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
	session.ApplicationName = session.Application.Name
	session.Status = SessionStatusStarting
	if session.Logs == nil {
		session.Logs = []Log{}
	}
	if session.CommandsLogs == nil {
		session.CommandsLogs = []string{}
	}
	if len(session.Variables) == 0 {
		session.Variables = make(map[string]string)
	}
	return session
}

func (session *Session) LogCritical(message string) {
	session.Logs = append(
		session.Logs,
		NewLog(message, LogTypeCritical),
	)
}

func (session *Session) LogError(message string) {
	session.Logs = append(
		session.Logs,
		NewLog(message, LogTypeError),
	)
}

func (session *Session) LogWarn(message string) {
	session.Logs = append(
		session.Logs,
		NewLog(message, LogTypeWarn),
	)
}

func (session *Session) LogInfo(message string) {
	session.Logs = append(
		session.Logs,
		NewLog(message, LogTypeInfo),
	)
}

func (session *Session) LogDebug(message string) {
	session.Logs = append(
		session.Logs,
		NewLog(message, LogTypeDebug),
	)
}

func (session *Session) LogTrace(message string) {
	session.Logs = append(
		session.Logs,
		NewLog(message, LogTypeTrace),
	)
}

func (session *Session) LogStdin(message string) {
	session.Logs = append(
		session.Logs,
		NewLog(message, LogTypeStdin),
	)
}

func (session *Session) LogStdout(message string) {
	session.Logs = append(
		session.Logs,
		NewLog(message, LogTypeStdout),
	)
}

func (session *Session) LogStderr(message string) {
	session.Logs = append(
		session.Logs,
		NewLog(message, LogTypeStderr),
	)
}

func (session *Session) MarkAsBeingRequested() {
	// Refreshes the inactiveAt field every time someone makes a request to this session
	session.InactiveAt = time.Now().Add(time.Second * time.Duration(session.Application.Recycle.InactivityTimeout))
	session.MaxAge = session.Application.Recycle.InactivityTimeout
}
