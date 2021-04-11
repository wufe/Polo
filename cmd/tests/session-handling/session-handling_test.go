package session_handling

import (
	"testing"

	"github.com/wufe/polo/internal/tests"
	"github.com/wufe/polo/pkg/models"
)

func Test_SessionHandling(t *testing.T) {

	tests.Fixture(&models.RootConfiguration{
		ApplicationConfigurations: []*models.ApplicationConfiguration{
			&models.ApplicationConfiguration{
				SharedConfiguration: models.SharedConfiguration{
					Remote: "https://git.example.com",
					Commands: models.Commands{
						Start: []models.Command{},
						Stop:  []models.Command{},
					},
				},
				Name:      "Test",
				IsDefault: true,
			},
		},
	})

}
