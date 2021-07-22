package application_init

import (
	"testing"

	"github.com/wufe/polo/internal/tests"
	"github.com/wufe/polo/internal/tests/events_assertions"
	"github.com/wufe/polo/pkg/models"
)

func Test_ApplicationInit(t *testing.T) {

	di := tests.Fixture(nil, models.BuildApplicationConfiguration("Test_ApplicationInit").
		WithRemote("https://github.com/wufe/polo-testserver"))
	applications := di.GetApplications()
	firstApplicationBus := applications[0].GetEventBus()

	events_assertions.AssertApplicationGetsInitializedAndFetched(firstApplicationBus.GetChan(), t)
}
