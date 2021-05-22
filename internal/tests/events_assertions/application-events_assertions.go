package events_assertions

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/wufe/polo/pkg/models"

	aurora "github.com/logrusorgru/aurora/v3"
)

func AssertApplicationEvents(
	ch <-chan models.ApplicationEvent,
	events []models.ApplicationEventType,
	t *testing.T,
	timeout time.Duration,
) []models.ApplicationEvent {

	stringifiedExpectedEventsSlice := []string{}
	for _, ev := range events {
		stringifiedExpectedEventsSlice = append(stringifiedExpectedEventsSlice, ev.String())
	}
	stringifiedExpectedEvents := strings.Join(stringifiedExpectedEventsSlice, ", ")

	lastFoundIndex := -1
	stringifiedGotEventsSlice := []string{}

	timeoutFired := false

	gotEvents := []models.ApplicationEvent{}

L:
	for {
		select {
		case ev, ok := <-ch:
			if !ok {
				break L
			}
			fmt.Printf("[APP_EVENT]: %s\n", ev.EventType)
			gotEvents = append(gotEvents, ev)
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
		t.Error(aurora.Sprintf(aurora.Red("expected application events to be:\n%s,\nbut timeout fired and got:\n%s"), stringifiedExpectedEvents, stringifiedGotEvents))
	} else {
		if lastFoundIndex < len(events)-1 {
			stringifiedGotEvents := strings.Join(stringifiedGotEventsSlice, ", ")
			t.Error(aurora.Sprintf(aurora.Red("expected application events to be:\n%s,\nbut got:\n%s instead"), stringifiedExpectedEvents, stringifiedGotEvents))
		}
	}
	return gotEvents
}

func AssertSessionEvents(
	ch <-chan models.SessionBuildEvent,
	events []models.SessionEventType,
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
			fmt.Printf("[SESSION_EVENT]: %s\n", ev.EventType)
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
		t.Errorf("expected session events to be %s, but timeout fired and got %s events", stringifiedExpectedEvents, stringifiedGotEvents)
	} else {
		if lastFoundIndex < len(events)-1 {
			stringifiedGotEvents := strings.Join(stringifiedGotEventsSlice, ", ")
			t.Errorf("expected session events to be %s, but got %s instead", stringifiedExpectedEvents, stringifiedGotEvents)
		}
	}
}
