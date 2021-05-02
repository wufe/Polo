package application_init

import (
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/wufe/polo/internal/tests"
	"github.com/wufe/polo/internal/tests/events_assertions"
	"github.com/wufe/polo/pkg/models"
)

func Test_ApplicationInit(t *testing.T) {

	log.SetLevel(log.PanicLevel)

	di := tests.Fixture(&models.ApplicationConfiguration{
		SharedConfiguration: models.SharedConfiguration{
			Remote: "https://github.com/wufe/polo-testserver",
			Commands: models.Commands{
				Start: []models.Command{},
				Stop:  []models.Command{},
			},
		},
		Name:      "Test_ApplicationInit",
		IsDefault: true,
	}, nil)
	applications := di.GetApplications()
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
