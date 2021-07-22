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
) ([]models.ApplicationEventType, bool) {

	lastFoundIndex := -1
	timeoutFired := false
	gotEvents := []models.ApplicationEventType{}
	succeded := true

L:
	for {
		select {
		case ev, ok := <-ch:
			if !ok {
				break L
			}
			fmt.Println(aurora.Sprintf(aurora.Yellow("[APP_EVENT]: %s"), ev.EventType))
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

	stringifiedGotEvents := stringifyApplicationEvents(gotEvents)
	stringifiedExpectedEvents := stringifyApplicationEvents(events)

	if timeoutFired {
		t.Error(aurora.Sprintf(aurora.Red("expected application events to be:\n%s,\nbut timeout fired and got:\n%s"), stringifiedExpectedEvents, stringifiedGotEvents))
		succeded = false
	} else if lastFoundIndex < len(events)-1 {
		t.Error(aurora.Sprintf(aurora.Red("expected application events to be:\n%s,\nbut got:\n%s instead"), stringifiedExpectedEvents, stringifiedGotEvents))
		succeded = false
	}
	return gotEvents, succeded
}

func AssertConcurrentApplicationEvents(
	ch <-chan models.ApplicationEvent,
	events []models.ApplicationEventType,
	t *testing.T,
	timeout time.Duration,
) ([]models.ApplicationEventType, bool) {

	mismatching := false
	timeoutFired := false
	gotEvents := []models.ApplicationEventType{}
	succeded := true

	eventsMap := make(map[models.ApplicationEventType]int)
	for _, e := range events {
		if n, ok := eventsMap[e]; ok {
			eventsMap[e] = n + 1
		} else {
			eventsMap[e] = 1
		}
	}

L:
	for {
		select {
		case ev, ok := <-ch:
			if !ok {
				break L
			}
			fmt.Println(aurora.Sprintf(aurora.Yellow("[APP_EVENT]: %s"), ev.EventType))
			gotEvents = append(gotEvents, ev.EventType)
			if available, accepted := eventsMap[ev.EventType]; !accepted || available == 0 {
				mismatching = true
				break L
			}
			eventsMap[ev.EventType] = eventsMap[ev.EventType] - 1
			if len(gotEvents) == len(events) {
				break L
			}
		case <-time.After(timeout):
			timeoutFired = true
			break L
		}
	}

	stringifiedExpectedEvents := stringifyApplicationEvents(events)
	stringifiedGotEvents := stringifyApplicationEvents(gotEvents)
	if timeoutFired {
		t.Error(aurora.Sprintf(aurora.Red("expected application events to be one of:\n%s,\nbut timeout fired and got:\n%s"), stringifiedExpectedEvents, stringifiedGotEvents))
		succeded = false
	} else if mismatching {
		t.Error(aurora.Sprintf(aurora.Red("expected application events to be one of:\n%s,\nbut got:\n%s instead"), stringifiedExpectedEvents, stringifiedGotEvents))
		succeded = false
	}
	return gotEvents, succeded
}

func stringifyApplicationEvents(events []models.ApplicationEventType) string {
	stringifiedExpectedEventsSlice := []string{}
	for _, ev := range events {
		stringifiedExpectedEventsSlice = append(stringifiedExpectedEventsSlice, ev.String())
	}
	return strings.Join(stringifiedExpectedEventsSlice, ", ")
}
