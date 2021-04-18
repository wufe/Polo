package models

import (
	"github.com/wufe/polo/pkg/background/communication"
	"github.com/wufe/polo/pkg/utils"
)

type SessionBuilder struct {
	mutexBuilder  utils.MutexBuilder
	pubSubBuilder *communication.PubSubBuilder
}

func NewSessionBuilder(mutexBuilder utils.MutexBuilder, pubSubBuilder *communication.PubSubBuilder) *SessionBuilder {
	return &SessionBuilder{
		mutexBuilder:  mutexBuilder,
		pubSubBuilder: pubSubBuilder,
	}
}

func (b *SessionBuilder) Build(session *Session) *Session {
	return newSession(session, b.mutexBuilder, b.pubSubBuilder)
}
