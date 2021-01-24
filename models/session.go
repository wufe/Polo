package models

import (
	"context"
	"time"
)

const (
	SessionStatusStarting    SessionStatus = "starting"
	SessionStatusStarted     SessionStatus = "started"
	SessionStatusStartFailed SessionStatus = "start_failed"
)

type SessionStatus string

type Session struct {
	UUID     string             `json:"uuid"`
	Name     string             `json:"name"`
	Target   string             `json:"target"`
	Port     int                `json:"port"`
	Service  *Service           `json:"service"`
	Status   SessionStatus      `json:"status"`
	Logs     []string           `json:"logs"`
	Checkout string             `json:"checkout"` // The object to be checked out (branch/tag/commit id)
	Done     chan struct{}      `json:"-"`
	Context  context.Context    `json:"-"`
	cancel   context.CancelFunc `json:"-"`
}

func NewSession(
	ctx context.Context,
	session *Session,
) *Session {
	sessionCtx, sessionCancel := context.WithTimeout(ctx, time.Second*time.Duration(session.Service.Healthcheck.RetryTimeout))
	session.Context = sessionCtx
	session.cancel = sessionCancel
	return session
}

func (session *Session) Cancel() {
	session.cancel()
}
