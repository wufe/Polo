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
	Logs         []string          `json:"logs"`
	Checkout     string            `json:"checkout"` // The object to be checked out (branch/tag/commit id)
	Done         chan struct{}     `json:"-"`
	InactiveAt   time.Time         `json:"inactiveAt"`
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
