package net

import (
	"github.com/phayes/freeport"
	log "github.com/sirupsen/logrus"
	"github.com/wufe/polo/pkg/models"
)

type PortRetriever interface {
	GetFreePort(portConfiguration models.PortConfiguration) (int, error)
}

type portRetrieverImpl struct{}

func NewPortRetriever() PortRetriever {
	return &portRetrieverImpl{}
}

func (r *portRetrieverImpl) GetFreePort(portConfiguration models.PortConfiguration) (int, error) {
	log.Trace("Getting a free port")
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
