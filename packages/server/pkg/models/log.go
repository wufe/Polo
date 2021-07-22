package models

import (
	"time"

	"github.com/google/uuid"
)

const (
	LogTypeTrace    LogType = "trace"
	LogTypeDebug    LogType = "debug"
	LogTypeInfo     LogType = "info"
	LogTypeWarn     LogType = "warn"
	LogTypeError    LogType = "error"
	LogTypeCritical LogType = "critical"
)

type Log struct {
	When    time.Time `json:"when"`
	UUID    string    `json:"uuid"`
	Type    LogType   `json:"type"`
	Message string    `json:"message"`
}

type LogType string

func NewLog(message string, logType LogType) Log {
	return Log{
		When:    time.Now(),
		UUID:    uuid.NewString(),
		Message: message,
		Type:    logType,
	}
}
