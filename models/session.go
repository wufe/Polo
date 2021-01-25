package models

import (
	"time"
)

const (
	SessionStatusStarting    SessionStatus = "starting"
	SessionStatusStarted     SessionStatus = "started"
	SessionStatusStartFailed SessionStatus = "start_failed"
	SessionStatusStopping    SessionStatus = "stopping"
)

type SessionStatus string

type Session struct {
	UUID       string        `json:"uuid"`
	Name       string        `json:"name"`
	Target     string        `json:"target"`
	Port       int           `json:"port"`
	Service    *Service      `json:"service"`
	Status     SessionStatus `json:"status"`
	Logs       []string      `json:"logs"`
	Checkout   string        `json:"checkout"` // The object to be checked out (branch/tag/commit id)
	Done       chan struct{} `json:"-"`
	InactiveAt time.Time     `json:"inactiveAt"`
	Folder     string        `json:"folder"`
}

func NewSession(
	session *Session,
) *Session {
	return session
}
