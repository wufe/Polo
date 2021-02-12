package mappers

import (
	"github.com/wufe/polo/pkg/models"
	"github.com/wufe/polo/pkg/services/output"
)

func MapSession(model *models.Session) *output.Session {
	if model == nil {
		return nil
	}
	model.Lock()
	defer model.Unlock()
	return &output.Session{
		UUID:            model.UUID,
		Name:            model.Name,
		Target:          model.Target,
		Port:            model.Port,
		ApplicationName: model.ApplicationName,
		Status:          string(model.Status),
		CommitID:        model.CommitID,
		Checkout:        model.Checkout,
		MaxAge:          model.MaxAge,
		Folder:          model.Folder,
		CommandsLogs:    model.CommandsLogs,
		Variables:       model.Variables,
		Logs:            MapSessionLogs(model.Logs),
		Metrics:         MapMetrics(model.Metrics),
	}
}

func MapSessions(models []*models.Session) []output.Session {
	ret := []output.Session{}
	for _, s := range models {
		ret = append(ret, *MapSession(s))
	}
	return ret
}
