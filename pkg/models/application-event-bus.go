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
)

type ApplicationEventType string

func (t ApplicationEventType) String() string {
	return string(t)
}

type ApplicationEvent struct {
	EventType   ApplicationEventType
	Application *Application
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

func (b *ApplicationEventBus) PublishEvent(eventType ApplicationEventType, application *Application) {
	b.pubSub.Publish("application", ApplicationEvent{
		EventType:   eventType,
		Application: application,
	})
}

func (b *ApplicationEventBus) Close() {
	b.pubSub.Close()
}

func (b *ApplicationEventBus) GetHistory() []ApplicationEvent {
	history := b.pubSub.GetHistory("application")
	return b.convertHistoryEntries(history)
}

func (b *ApplicationEventBus) convertHistoryEntries(history *communication.PubSubHistory) []ApplicationEvent {
	applicationEvents := []ApplicationEvent{}
	for _, rawEvent := range history.GetEntries() {
		applicationEvent, ok := rawEvent.(ApplicationEvent)
		if ok {
			applicationEvents = append(applicationEvents, applicationEvent)
		}
	}
	return applicationEvents
}
