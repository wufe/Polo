package services

import uuid "github.com/iris-contrib/go.uuid"

type AliasingService struct{}

func NewAliasingService() *AliasingService {
	return &AliasingService{}
}

func (a *AliasingService) Next(exclude []string) string {
	return uuid.Must(uuid.NewV1()).String()
}
