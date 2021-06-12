package models

import "github.com/wufe/polo/pkg/models/output"

func mapApplicationNotifications(models []ApplicationNotification) []output.ApplicationNotification {
	if models == nil {
		return []output.ApplicationNotification{}
	}
	ret := make([]output.ApplicationNotification, 0, len(models))
	for _, notification := range models {
		ret = append(ret, notification.ToOutput())
	}
	return ret
}

func mapApplicationError(err ApplicationNotification) output.ApplicationNotification {
	return output.ApplicationNotification{
		UUID:        err.UUID,
		Type:        string(err.Type),
		Permanent:   err.Permanent,
		Level:       string(err.Level),
		Description: err.Description,
		CreatedAt:   err.CreatedAt,
	}
}
