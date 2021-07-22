package events_assertions

import (
	"fmt"
	"strings"
	"testing"
	"time"

	aurora "github.com/logrusorgru/aurora/v3"
	"github.com/wufe/polo/pkg/models"
)

func AssertSessionEvents(
	ch <-chan models.SessionBuildEvent,
	events []models.SessionEventType,
	t *testing.T,
	timeout time.Duration,
) {

	lastFoundIndex := -1
	timeoutFired := false
	gotEvents := []models.SessionEventType{}

L:
	for {
		select {
		case ev, ok := <-ch:
			if !ok {
				break L
			}
			fmt.Println(aurora.Sprintf(aurora.Cyan("[SESSION_EVENT]: %s"), ev.EventType))
			gotEvents = append(gotEvents, ev.EventType)
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

	stringifiedExpectedEvents := stringifySessionEvents(events)
	stringifiedGotEvents := stringifySessionEvents(gotEvents)
	if timeoutFired {
		t.Error(aurora.Sprintf(aurora.Red("expected application events to be:\n%s,\nbut timeout fired and got:\n%s"), stringifiedExpectedEvents, stringifiedGotEvents))
	} else {
		if lastFoundIndex < len(events)-1 {
			t.Error(aurora.Sprintf(aurora.Red("expected application events to be:\n%s,\nbut got:\n%s instead"), stringifiedExpectedEvents, stringifiedGotEvents))
		}
	}
}

func stringifySessionEvents(events []models.SessionEventType) string {
	stringifiedExpectedEventsSlice := []string{}
	for _, ev := range events {
		stringifiedExpectedEventsSlice = append(stringifiedExpectedEventsSlice, ev.String())
	}
	return strings.Join(stringifiedExpectedEventsSlice, ", ")
}
