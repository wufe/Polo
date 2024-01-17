package models

import (
	tilt_output_models "github.com/wufe/polo/pkg/integrations/tilt/models/output"
)

type Session struct {
	Tilt tilt_output_models.Session `json:"tilt"`
}
