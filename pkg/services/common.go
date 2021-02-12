package services

import "errors"

var (
	ErrApplicationNotFound error = errors.New("Application not found")
	ErrSessionNotFound     error = errors.New("Session not found")
	ErrSessionIsNotAlive   error = errors.New("Session is not alive")
)
