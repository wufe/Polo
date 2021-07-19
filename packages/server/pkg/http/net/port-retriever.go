package net

import (
	"github.com/phayes/freeport"
	"github.com/wufe/polo/pkg/logging"
	"github.com/wufe/polo/pkg/models"
)

type PortRetriever interface {
	GetFreePort(portConfiguration models.PortConfiguration) (int, error)
}

type portRetrieverImpl struct {
	log logging.Logger
}

func NewPortRetriever(logger logging.Logger) PortRetriever {
	return &portRetrieverImpl{
		log: logger,
	}
}

func (r *portRetrieverImpl) GetFreePort(portConfiguration models.PortConfiguration) (int, error) {
	r.log.Trace("Getting a free port")
	foundPort := 0
L:
	for foundPort == 0 {
		freePort, err := freeport.GetFreePort()
		if err != nil {
			return 0, err
		}
		for _, port := range portConfiguration.Except {
			if freePort == port {
				continue L
			}
		}
		foundPort = freePort
	}
	return foundPort, nil
}
