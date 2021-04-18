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
	pubSub     *communication.PubSub
	eventsChan chan ApplicationEvent
}

func NewApplicationEventBus(pubSubBuilder *communication.PubSubBuilder) *ApplicationEventBus {
	pubSub := pubSubBuilder.Build()
	eventBus := &ApplicationEventBus{
		pubSub:     pubSub,
		eventsChan: make(chan ApplicationEvent),
	}
	eventBus.convertEvents()
	return eventBus
}

func (b *ApplicationEventBus) convertEvents() {
	ch := b.pubSub.Subscribe("application")
	go func() {
		for {
			rawEvent, ok := <-ch
			if !ok {
				return
			}
			go func() {
				applicationEvent, ok := rawEvent.(ApplicationEvent)
				if ok {
					b.eventsChan <- applicationEvent
				}
			}()
		}
	}()
}

func (b *ApplicationEventBus) GetChan() <-chan ApplicationEvent {
	return b.eventsChan
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
	applicationEvents := []ApplicationEvent{}
	for _, rawEvent := range history.GetEntries() {
		applicationEvent, ok := rawEvent.(ApplicationEvent)
		if ok {
			applicationEvents = append(applicationEvents, applicationEvent)
		}
	}
	return applicationEvents
}
