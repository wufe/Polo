package models

import "github.com/wufe/polo/pkg/models/output"

func mapApplicationErrors(models []ApplicationError) []output.ApplicationError {
	if models == nil {
		return []output.ApplicationError{}
	}
	ret := make([]output.ApplicationError, 0, len(models))
	for _, err := range models {
		ret = append(ret, err.ToOutput())
	}
	return ret
}

func mapApplicationError(err ApplicationError) output.ApplicationError {
	return output.ApplicationError{
		UUID:        err.UUID,
		Type:        string(err.Type),
		Description: err.Description,
		CreatedAt:   err.CreatedAt,
	}
}
