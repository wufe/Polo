package integrations

import (
	"github.com/wufe/polo/pkg/integrations/models"
	"github.com/wufe/polo/pkg/integrations/tilt"
)

func ParseSessionCommandOutput(model *models.Session, output string) *models.Session {
	model.Tilt = tilt.ParseSessionCommandOutput(model.Tilt, output)
	return model
}
