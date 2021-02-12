package mappers

import (
	"github.com/wufe/polo/pkg/models"
	"github.com/wufe/polo/pkg/services/output"
)

func MapSessionLog(log models.Log) output.SessionLog {
	return output.SessionLog{
		When:    log.When,
		UUID:    log.UUID,
		Type:    string(log.Type),
		Message: log.Message,
	}
}

func MapSessionLogs(logs []models.Log) []output.SessionLog {
	ret := []output.SessionLog{}
	for _, log := range logs {
		ret = append(ret, MapSessionLog(log))
	}
	return ret
}
