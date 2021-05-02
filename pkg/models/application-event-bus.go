package models

import (
	"github.com/wufe/polo/pkg/background/communication"
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
	pubSub *communication.PubSub
}

func NewApplicationEventBus(pubSubBuilder *communication.PubSubBuilder) *ApplicationEventBus {
	pubSub := pubSubBuilder.Build()
	eventBus := &ApplicationEventBus{
		pubSub: pubSub,
	}
	return eventBus
}

func (b *ApplicationEventBus) GetChan() <-chan ApplicationEvent {
	sourceCh, history := b.pubSub.Subscribe("application")
	destCh := make(chan ApplicationEvent)
	go func() {
		for _, pastEv := range b.convertHistoryEntries(history) {
			destCh <- pastEv
		}
		for {
			ev, ok := <-sourceCh
			if !ok {
				close(destCh)
				return
			}
			appEv, ok := ev.(ApplicationEvent)
			if !ok {
				continue
			}
			destCh <- appEv
		}
	}()
	return destCh
}

func (b *ApplicationEventBus) PublishEvent(eventType ApplicationEventType, application *Application, payloadObjects ...interface{}) {

	var payload interface{} = nil
	if len(payloadObjects) == 1 {
		payload = payloadObjects
	} else {
		payload = payloadObjects
	}

	b.pubSub.Publish("application", ApplicationEvent{
		EventType:    eventType,
		Application:  application,
		EventPayload: payload,
	})
}

func (b *ApplicationEventBus) Close() {
	b.pubSub.Close()
}

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
