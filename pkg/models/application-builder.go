package models

import "github.com/wufe/polo/pkg/utils"

type ApplicationBuilder struct {
	mutexBuilder utils.MutexBuilder
}

func NewApplicationBuilder(mutexBuilder utils.MutexBuilder) *ApplicationBuilder {
	return &ApplicationBuilder{
		mutexBuilder: mutexBuilder,
	}
}

func (b *ApplicationBuilder) Build(configuration *ApplicationConfiguration, filename string) (*Application, error) {
	return newApplication(configuration, filename, b.mutexBuilder)
}

func (b *ApplicationBuilder) BuildConfiguration(configuration *ApplicationConfiguration) (*ApplicationConfiguration, error) {
	return NewApplicationConfiguration(configuration, b.mutexBuilder)
}
