package models

import "time"

const (
	DiagonsticsActionReplacement DiagnosticsAction = "replacement"
)

type DiagnosticsAction string

func (diagnosticsAction DiagnosticsAction) String() string {
	return string(diagnosticsAction)
}

type DiagnosticsData struct {
	Action DiagnosticsAction
	When   time.Time
	Field  string
	Value  interface{}
}

type PrevNextDiagnosticsValue struct {
	Previous string
	Next     string
}
