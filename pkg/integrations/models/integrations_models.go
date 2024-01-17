package models

import (
	tilt_models "github.com/wufe/polo/pkg/integrations/tilt/models"
)

type Session struct {
	Tilt tilt_models.Session `json:"tilt"`
}

func NewSession(previousStoredModel *Session) *Session {
	if previousStoredModel == nil {
		return &Session{
			Tilt: *tilt_models.NewSession(nil),
		}
	}
	return &Session{
		Tilt: *tilt_models.NewSession(&previousStoredModel.Tilt),
	}
}
