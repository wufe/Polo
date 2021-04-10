package utils

import (
	"sync"

	"github.com/sasha-s/go-deadlock"
)

type RWLocker interface {
	Lock()
	Unlock()
	RLock()
	RUnlock()
}

func GetMutex(environment Environment) RWLocker {
	if environment.IsDev() && environment.IsDebugRace() {
		return &deadlock.RWMutex{}
	} else {
		return &sync.RWMutex{}
	}
}
