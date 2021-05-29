package events_assertions

import (
	"testing"
	"time"

	"github.com/wufe/polo/pkg/models"
)

func AssertApplicationGetsFetched(appChan <-chan models.ApplicationEvent, t *testing.T) {
	AssertApplicationEvents(
		appChan,
		[]models.ApplicationEventType{
			models.ApplicationEventTypeFetchStarted,
			models.ApplicationEventTypeFetchCompleted,
		},
		t,
		2*time.Second,
	)
}

func AssertApplicationSessionSucceeded(appChan <-chan models.ApplicationEvent, t *testing.T) {
	AssertApplicationEvents(
		appChan,
		[]models.ApplicationEventType{
			models.ApplicationEventTypeSessionBuild,
			models.ApplicationEventTypeSessionBuildSucceeded,
		},
		t,
		2*time.Second,
	)
}

func AssertApplicationSessionBuildFails4Times(appChan <-chan models.ApplicationEvent, t *testing.T) {
	_, succeded := AssertApplicationEvents(
		appChan,
		[]models.ApplicationEventType{
			models.ApplicationEventTypeSessionBuild,
			models.ApplicationEventTypeSessionBuildFailed,
			models.ApplicationEventTypeSessionBuild,
		},
		t,
		3*time.Second,
	)

	if succeded {
		_, succeded = ConcurrentlyAssertCleanedAndFailed(appChan, t)
	}

	if succeded {
		_, succeded = AssertApplicationEvents(
			appChan,
			[]models.ApplicationEventType{
				models.ApplicationEventTypeSessionBuild,
			},
			t,
			2*time.Second,
		)
	}

	if succeded {
		_, succeded = ConcurrentlyAssertCleanedAndFailed(appChan, t)
	}

	if succeded {
		_, succeded = AssertApplicationEvents(
			appChan,
			[]models.ApplicationEventType{
				models.ApplicationEventTypeSessionBuild,
			},
			t,
			2*time.Second,
		)
	}
	if succeded {
		_, succeded = ConcurrentlyAssertCleanedFailedAndCleaned(appChan, t)
	}
}

func AssertApplicationGetsInitializedAndFetched(appChan <-chan models.ApplicationEvent, t *testing.T) {
	AssertApplicationEvents(
		appChan,
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

func AssertApplicationGetsFetchedWithHotSwap(appChan <-chan models.ApplicationEvent, t *testing.T) {
	AssertApplicationEvents(
		appChan,
		[]models.ApplicationEventType{
			models.ApplicationEventTypeFetchStarted,
			models.ApplicationEventTypeHotSwap,
			models.ApplicationEventTypeSessionBuild,
		},
		t,
		2*time.Second,
	)
}

func AssertApplicationGetsFetchedWithHotSwapAndFailingBuild(appChan <-chan models.ApplicationEvent, t *testing.T) {
	AssertApplicationEvents(
		appChan,
		[]models.ApplicationEventType{
			models.ApplicationEventTypeFetchStarted,
			models.ApplicationEventTypeHotSwap,
			models.ApplicationEventTypeSessionBuild,
			models.ApplicationEventTypeFetchCompleted,
			models.ApplicationEventTypeSessionBuildFailed,
		},
		t,
		2*time.Second,
	)
}

func AssertApplicationGetsFetchedWithHotSwapAndFailingBuildWith3Retries(appChan <-chan models.ApplicationEvent, t *testing.T) {
	_, succeded := AssertApplicationEvents(
		appChan,
		[]models.ApplicationEventType{
			models.ApplicationEventTypeFetchStarted,
			models.ApplicationEventTypeHotSwap,
			models.ApplicationEventTypeSessionBuild,
			models.ApplicationEventTypeFetchCompleted,
			models.ApplicationEventTypeSessionBuildFailed,
			models.ApplicationEventTypeSessionBuild,
		},
		t,
		2*time.Second,
	)
	if succeded {
		_, succeded = ConcurrentlyAssertCleanedAndFailed(appChan, t)
	}
	if succeded {
		_, succeded = AssertApplicationEvents(
			appChan,
			[]models.ApplicationEventType{
				models.ApplicationEventTypeSessionBuild,
			},
			t,
			2*time.Second,
		)
	}
	if succeded {
		_, succeded = ConcurrentlyAssertCleanedAndFailed(appChan, t)
	}
	if succeded {
		_, succeded = AssertApplicationEvents(
			appChan,
			[]models.ApplicationEventType{
				models.ApplicationEventTypeSessionBuild,
			},
			t,
			2*time.Second,
		)
	}
	if succeded {
		_, succeded = ConcurrentlyAssertCleanedFailedAndCleaned(appChan, t)
	}
}

func ConcurrentlyAssertCleanedAndFailed(appChan <-chan models.ApplicationEvent, t *testing.T) ([]models.ApplicationEventType, bool) {
	return AssertConcurrentApplicationEvents(
		appChan,
		[]models.ApplicationEventType{
			models.ApplicationEventTypeSessionCleaned,
			models.ApplicationEventTypeSessionBuildFailed,
		},
		t,
		2*time.Second,
	)
}

func ConcurrentlyAssertCleanedFailedAndCleaned(appChan <-chan models.ApplicationEvent, t *testing.T) ([]models.ApplicationEventType, bool) {
	return AssertConcurrentApplicationEvents(
		appChan,
		[]models.ApplicationEventType{
			models.ApplicationEventTypeSessionCleaned,
			models.ApplicationEventTypeSessionBuildFailed,
			models.ApplicationEventTypeSessionCleaned,
		},
		t,
		2*time.Second,
	)
}
