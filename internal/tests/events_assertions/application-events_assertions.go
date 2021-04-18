package events_assertions

import (
	"strings"
	"testing"
	"time"

	"github.com/wufe/polo/pkg/models"
)

func AssertApplicationEvents(
	ch <-chan models.ApplicationEvent,
	events []models.ApplicationEventType,
	t *testing.T,
	timeout time.Duration,
) {

	stringifiedExpectedEventsSlice := []string{}
	for _, ev := range events {
		stringifiedExpectedEventsSlice = append(stringifiedExpectedEventsSlice, ev.String())
	}
	stringifiedExpectedEvents := strings.Join(stringifiedExpectedEventsSlice, ", ")

	lastFoundIndex := -1
	stringifiedGotEventsSlice := []string{}

	timeoutFired := false

L:
	for {
		select {
		case ev, ok := <-ch:
			if !ok {
				break L
			}
			stringifiedGotEventsSlice = append(stringifiedGotEventsSlice, ev.EventType.String())
			if ev.EventType == events[lastFoundIndex+1] {
				lastFoundIndex++
				if lastFoundIndex == len(events)-1 {
					break L
				}
			} else {
				break L
			}
		case <-time.After(timeout):
			timeoutFired = true
			break L
		}
	}

	if timeoutFired {
		stringifiedGotEvents := strings.Join(stringifiedGotEventsSlice, ", ")
		t.Errorf("expected application events to be %s, but timeout fired and got %s events", stringifiedExpectedEvents, stringifiedGotEvents)
	} else {
		if lastFoundIndex < len(events)-1 {
			stringifiedGotEvents := strings.Join(stringifiedGotEventsSlice, ", ")
			t.Errorf("expected application events to be %s, but got %s instead", stringifiedExpectedEvents, stringifiedGotEvents)
		}
	}
}
