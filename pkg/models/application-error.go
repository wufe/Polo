package models

import (
	"time"

	"github.com/wufe/polo/pkg/models/output"
)

const (
	ApplicationErrorTypeGitCredentials ApplicationErrorType = "git_credentials_error"
)

type ApplicationErrorType string

type ApplicationError struct {
	UUID        string
	Type        ApplicationErrorType
	Description string
	CreatedAt   time.Time
}

func (e *ApplicationError) ToOutput() output.ApplicationError {
	return mapApplicationError(*e)
}
