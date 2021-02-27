package mappers

import (
	"github.com/wufe/polo/pkg/models"
	"github.com/wufe/polo/pkg/services/output"
)

func MapSession(model *models.Session) *output.Session {
	if model == nil {
		return nil
	}
	conf := model.GetConfiguration()
	model.Lock()
	session := &output.Session{
		UUID:              model.UUID,
		Name:              model.Name,
		Target:            model.Target,
		Port:              model.Port,
		ApplicationName:   model.ApplicationName,
		Status:            string(model.Status),
		CommitID:          model.CommitID,
		CommitMessage:     model.Commit.Message,
		CommitAuthorName:  model.Commit.Author.Name,
		CommitAuthorEmail: model.Commit.Author.Email,
		CommitDate:        model.Commit.Author.When,
		Checkout:          model.Checkout,
		MaxAge:            model.MaxAge,
		Folder:            model.Folder,
		Variables:         model.Variables,
		Logs:              MapSessionLogs(model.Logs),
		Metrics:           MapMetrics(model.Metrics),
		Configuration:     MapConfiguration(conf),
	}
	model.Unlock()
	session.ReplacesSession = MapReplaces(model.Replaces())
	return session
}

func MapSessions(models []*models.Session) []output.Session {
	ret := []output.Session{}
	for _, s := range models {
		ret = append(ret, *MapSession(s))
	}
	return ret
}

func MapReplaces(model *models.Session) string {
	if model == nil {
		return ""
	}
	return model.UUID
}

func MapConfiguration(model models.ApplicationConfiguration) output.SessionConfiguration {
	return output.SessionConfiguration{
		IsDefault: model.IsDefault,
	}
}
