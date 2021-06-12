package models

import (
	"time"

	"github.com/wufe/polo/pkg/models/output"
)

const (
	ApplicationNotificationTypeGitClone ApplicationNotificationType = "git_clone_error"

	ApplicationNotificationLevelCritical ApplicationNotificationLevel = "critical"
)

type ApplicationNotificationType string
type ApplicationNotificationLevel string

type ApplicationNotification struct {
	UUID        string
	Type        ApplicationNotificationType
	Permanent   bool
	Level       ApplicationNotificationLevel
	Description string
	CreatedAt   time.Time
}

func (n *ApplicationNotification) ToOutput() output.ApplicationNotification {
	return mapApplicationError(*n)
}
