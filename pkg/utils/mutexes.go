package utils

import (
	"sync"

	"github.com/sasha-s/go-deadlock"
)

func GetMutex() sync.Locker {
	if IsDev() && IsDebugRace() {
		return &deadlock.Mutex{}
	} else {
		return &sync.Mutex{}
	}
}
