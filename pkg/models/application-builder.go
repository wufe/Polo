package models

import (
	"github.com/wufe/polo/pkg/logging"
	"github.com/wufe/polo/pkg/utils"
)

type ApplicationBuilder struct {
	mutexBuilder utils.MutexBuilder
	logger       logging.Logger
}

func NewApplicationBuilder(mutexBuilder utils.MutexBuilder, logger logging.Logger) *ApplicationBuilder {
	return &ApplicationBuilder{
		mutexBuilder: mutexBuilder,
		logger:       logger,
	}
}

func (b *ApplicationBuilder) Build(configuration *ApplicationConfiguration, filename string) (*Application, error) {
	return newApplication(configuration, filename, b.mutexBuilder, b.logger)
}

func (b *ApplicationBuilder) BuildConfiguration(configuration *ApplicationConfiguration) (*ApplicationConfiguration, error) {
	return NewApplicationConfiguration(configuration, b.mutexBuilder)
}
