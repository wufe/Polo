package models

import (
	log "github.com/sirupsen/logrus"
	"github.com/wufe/polo/pkg/utils"
)

type SessionBuilder struct {
	mutexBuilder utils.MutexBuilder
}

func NewSessionBuilder(mutexBuilder utils.MutexBuilder) *SessionBuilder {
	return &SessionBuilder{
		mutexBuilder: mutexBuilder,
	}
}

func (b *SessionBuilder) Build(session *Session) *Session {
	log.Trace("Building new session")
	return newSession(session, b.mutexBuilder)
}
