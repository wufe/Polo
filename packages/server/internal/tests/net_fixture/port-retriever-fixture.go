package net_fixture

import (
	"github.com/wufe/polo/pkg/models"
)

type portRetrieverFixtureImpl struct {
	port int
}

func NewPortRetrieverFixture() *portRetrieverFixtureImpl {
	return &portRetrieverFixtureImpl{}
}

func (r *portRetrieverFixtureImpl) SetFreePort(port int) {
	r.port = port
}

func (r *portRetrieverFixtureImpl) GetFreePort(portConfiguration models.PortConfiguration) (int, error) {
	return r.port, nil
}
