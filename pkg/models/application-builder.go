package models

import (
	"github.com/wufe/polo/pkg/background/communication"
	"github.com/wufe/polo/pkg/utils"
)

type ApplicationBuilder struct {
	mutexBuilder  utils.MutexBuilder
	pubSubBuilder *communication.PubSubBuilder
}

func NewApplicationBuilder(mutexBuilder utils.MutexBuilder, pubSubBuilder *communication.PubSubBuilder) *ApplicationBuilder {
	return &ApplicationBuilder{
		mutexBuilder:  mutexBuilder,
		pubSubBuilder: pubSubBuilder,
	}
}

func (b *ApplicationBuilder) Build(configuration *ApplicationConfiguration, filename string) (*Application, error) {
	return newApplication(configuration, filename, b.mutexBuilder, b.pubSubBuilder)
}

func (b *ApplicationBuilder) BuildConfiguration(configuration *ApplicationConfiguration) (*ApplicationConfiguration, error) {
	return NewApplicationConfiguration(configuration, b.mutexBuilder)
}
