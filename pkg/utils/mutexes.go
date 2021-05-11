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
	if (environment.IsDev() && environment.IsDebugRace()) || environment.IsDiagnostics() {
		return &deadlock.RWMutex{}
	} else if environment.IsTest() {
		return &deadlock.RWMutex{}
	} else {
		return &sync.RWMutex{}
	}
}

type MutexBuilder func() RWLocker
