package models

import (
	"github.com/wufe/polo/pkg/models/output"
)

func MapSession(model *Session) *output.Session {
	if model == nil {
		return nil
	}
	conf := model.GetConfiguration()
	status := MapSessionStatus(model)
	model.Lock()
	session := &output.Session{
		UUID:              model.UUID,
		Name:              model.Name,
		Target:            model.Target,
		Port:              model.Port,
		ApplicationName:   model.ApplicationName,
		CommitID:          model.CommitID,
		CommitMessage:     model.Commit.Message,
		CommitAuthorName:  model.Commit.Author.Name,
		CommitAuthorEmail: model.Commit.Author.Email,
		CommitDate:        model.Commit.Author.When,
		Checkout:          model.Checkout,
		Folder:            model.Folder,
		Variables:         model.Variables,
		Logs:              mapSessionLogs(model.Logs),
		Metrics:           mapMetrics(model.Metrics),
		Configuration:     mapConfiguration(conf),
		SessionStatus:     status,
	}
	model.Unlock()
	session.ReplacesSession = mapReplaces(model.Replaces())
	return session
}

func MapSessions(models []*Session) []output.Session {
	ret := []output.Session{}
	for _, s := range models {
		ret = append(ret, s.ToOutput())
	}
	return ret
}

func mapReplaces(model *Session) string {
	if model == nil {
		return ""
	}
	return model.UUID
}

func mapConfiguration(model ApplicationConfiguration) output.SessionConfiguration {
	return output.SessionConfiguration{
		IsDefault: model.IsDefault,
	}
}

// MapSessionStatus maps a session to a status output model
func MapSessionStatus(model *Session) output.SessionStatus {
	age := model.GetMaxAge()
	killReason := model.GetKillReason()
	replacedBy := model.GetReplacedBy()
	model.RLock()
	defer model.RUnlock()
	return output.SessionStatus{
		Status:     string(model.Status),
		Age:        age,
		KillReason: string(killReason),
		ReplacedBy: replacedBy,
	}
}

func MapSessionLog(log Log) output.SessionLog {
	return output.SessionLog{
		When:    log.When,
		UUID:    log.UUID,
		Type:    string(log.Type),
		Message: log.Message,
	}
}

func mapSessionLogs(logs []Log) []output.SessionLog {
	ret := []output.SessionLog{}
	for _, log := range logs {
		ret = append(ret, MapSessionLog(log))
	}
	return ret
}

func MapMetric(model Metric) output.Metric {
	return output.Metric{
		Object:   model.Object,
		Duration: int(model.Duration),
	}
}

func mapMetrics(models []Metric) []output.Metric {
	ret := []output.Metric{}
	for _, met := range models {
		ret = append(ret, output.Metric{
			Object:   met.Object,
			Duration: int(met.Duration),
		})
	}
	return ret
}
