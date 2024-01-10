package models

import (
	"github.com/wufe/polo/pkg/logging"
	"github.com/wufe/polo/pkg/utils"
)

type SessionBuilder struct {
	mutexBuilder  utils.MutexBuilder
	logger        logging.Logger
	configuration *RootConfiguration
}

func NewSessionBuilder(mutexBuilder utils.MutexBuilder, logger logging.Logger, configuration *RootConfiguration) *SessionBuilder {
	return &SessionBuilder{
		mutexBuilder:  mutexBuilder,
		logger:        logger,
		configuration: configuration,
	}
}

func (b *SessionBuilder) Build(session *Session) *Session {
	b.logger.Trace("Building new session")
	return newSession(session, b.mutexBuilder, b.logger, b.configuration.Global.FeaturesPreview.AdvancedTerminalOutput)
}
