package application_init

import (
	"testing"
	"time"

	"github.com/wufe/polo/internal/tests"
	"github.com/wufe/polo/internal/tests/events_assertions"
	"github.com/wufe/polo/pkg/models"
)

func Test_ApplicationInit(t *testing.T) {

	applications := tests.Fixture(&models.ApplicationConfiguration{
		SharedConfiguration: models.SharedConfiguration{
			Remote: "https://github.com/wufe/polo-testserver",
			Commands: models.Commands{
				Start: []models.Command{},
				Stop:  []models.Command{},
			},
		},
		Name:      "TestServer",
		IsDefault: true,
	}, nil)
	firstApplicationBus := applications[0].GetEventBus()

	events_assertions.AssertApplicationEvents(
		firstApplicationBus.GetChan(),
		[]models.ApplicationEventType{
			models.ApplicationEventTypeInitializationStarted,
			models.ApplicationEventTypeFetchStarted,
			models.ApplicationEventTypeFetchCompleted,
			models.ApplicationEventTypeInitializationCompleted,
		},
		t,
		10*time.Second,
	)
}
