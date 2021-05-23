package models

import (
	"github.com/asaskevich/EventBus"
	"github.com/wufe/polo/pkg/logging"
	"github.com/wufe/polo/pkg/utils"
)

const (
	ApplicationEventTypeNone                    ApplicationEventType = "none"
	ApplicationEventTypeInitializationStarted   ApplicationEventType = "initialization_started"
	ApplicationEventTypeInitializationCompleted ApplicationEventType = "initialization_completed"
	ApplicationEventTypeFetchStarted            ApplicationEventType = "fetch_started"
	ApplicationEventTypeFetchCompleted          ApplicationEventType = "fetch_completed"
	ApplicationEventTypeHotSwap                 ApplicationEventType = "hot_swap"
	ApplicationEventTypeAutoStart               ApplicationEventType = "auto_start"
	ApplicationEventTypeSessionBuild            ApplicationEventType = "session_build"
	ApplicationEventTypeSessionBuildFailed      ApplicationEventType = "session_build_failed"
	ApplicationEventTypeSessionBuildSucceeded   ApplicationEventType = "session_build_succeeded"
	ApplicationEventTypeSessionCleaned          ApplicationEventType = "session_cleaned"

	eventsBuffer = 100
)

type ApplicationEventType string

func (t ApplicationEventType) String() string {
	return string(t)
}

type ApplicationEvent struct {
	EventType    ApplicationEventType
	Application  *Application
	EventPayload interface{}
}

type ApplicationEventBus struct {
	utils.RWLocker
	bus     EventBus.Bus
	ch      chan ApplicationEvent
	history []ApplicationEvent
	log     logging.Logger
}

func NewApplicationEventBus(mutexBuilder utils.MutexBuilder, logger logging.Logger) *ApplicationEventBus {
	eventBus := &ApplicationEventBus{
		RWLocker: mutexBuilder(),
		bus:      EventBus.New(),
		ch:       make(chan ApplicationEvent, eventsBuffer),
		log:      logger,
	}
	eventBus.start()
	return eventBus
}

func (b *ApplicationEventBus) start() {
	history := []ApplicationEvent{}
	b.history = history

	eventsCount := 0

	b.bus.Subscribe("application", func(ev interface{}) {
		if appEv, ok := ev.(ApplicationEvent); ok {
			b.Lock()
			defer b.Unlock()
			history = append(history, appEv)
			b.ch <- appEv
			eventsCount++

			// Hack to prevent saturation of the receiving channel
			// if there are no listeners.
			if eventsCount >= eventsBuffer/2 {
			L:
				for {
					select {
					case <-b.ch:
					default:
						break L
					}
				}
				eventsCount = 0
			}
		}
	})
}

func (b *ApplicationEventBus) GetChan() <-chan ApplicationEvent {
	return b.ch
}

func (b *ApplicationEventBus) PublishEvent(eventType ApplicationEventType, application *Application, payloadObjects ...interface{}) {
	b.log.Tracef("Publishing event %q", eventType)
	var payload interface{} = nil
	if len(payloadObjects) == 1 {
		payload = payloadObjects
	} else {
		payload = payloadObjects
	}

	b.bus.Publish("application", ApplicationEvent{
		EventType:    eventType,
		Application:  application,
		EventPayload: payload,
	})
}

func (b *ApplicationEventBus) Close() {}

func (b *ApplicationEventBus) convertHistoryEntries(entries []interface{}) []ApplicationEvent {
	applicationEvents := []ApplicationEvent{}
	for _, rawEvent := range entries {
		applicationEvent, ok := rawEvent.(ApplicationEvent)
		if ok {
			applicationEvents = append(applicationEvents, applicationEvent)
		}
	}
	return applicationEvents
}
