package models

const (
	LogTypeTrace    LogType = "trace"
	LogTypeDebug    LogType = "debug"
	LogTypeInfo     LogType = "info"
	LogTypeWarn     LogType = "warn"
	LogTypeError    LogType = "error"
	LogTypeCritical LogType = "critical"
)

type Log struct {
	Type    LogType `json:"type"`
	Message string  `json:"message"`
}

type LogType string
