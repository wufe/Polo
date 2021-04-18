package models

import (
	"github.com/wufe/polo/pkg/background/communication"
)

const (
	SessionBuildEventTypeBuildStarted             SessionBuildEventType = "build_started"
	SessionBuildEventTypePreparingFolders         SessionBuildEventType = "preparing_folders"
	SessionBuildEventTypePreparingFoldersFailed   SessionBuildEventType = "preparing_folders_failed"
	SessionBuildEventTypeCommandsExecutionStarted SessionBuildEventType = "commands_execution_started"
	SessionBuildEventTypeCommandsExecutionFailed  SessionBuildEventType = "commands_execution_failed"
	SessionBuildEventTypeWarmupStarted            SessionBuildEventType = "warmup_started"
	SessionBuildEventTypeWarmupFailed             SessionBuildEventType = "warmup_failed"
	SessionBuildEventTypeHealthcheckStarted       SessionBuildEventType = "healthcheck_started"
	SessionBuildEventTypeHealthcheckFailed        SessionBuildEventType = "healthcheck_failed"
	SessionBuildEventTypeStarted                  SessionBuildEventType = "started"
)

type SessionBuildEventType string

func (sessionBuildEventType SessionBuildEventType) String() string {
	return string(sessionBuildEventType)
}

type SessionBuildEvent struct {
	EventType SessionBuildEventType
	Session   *Session
}

type SessionLifetimeEventBus struct {
	pubSub     *communication.PubSub
	eventsChan chan SessionBuildEvent
}

func NewSessionBuildEventBus(pubSubBuilder *communication.PubSubBuilder) *SessionLifetimeEventBus {
	pubSub := pubSubBuilder.Build()
	eventBus := &SessionLifetimeEventBus{
		pubSub:     pubSub,
		eventsChan: make(chan SessionBuildEvent),
	}
	eventBus.convertEvents()
	return eventBus
}

func (b *SessionLifetimeEventBus) convertEvents() {
	ch := b.pubSub.Subscribe("session")
	go func() {
		for {
			rawEvent, ok := <-ch
			if !ok {
				return
			}
			go func() {
				sessionEvent, ok := rawEvent.(SessionBuildEvent)
				if ok {
					b.eventsChan <- sessionEvent
				}
			}()
		}
	}()
}

func (b *SessionLifetimeEventBus) GetChan() <-chan SessionBuildEvent {
	return b.eventsChan
}

func (b *SessionLifetimeEventBus) PublishEvent(eventType SessionBuildEventType, session *Session) {
	b.pubSub.Publish("session", SessionBuildEvent{
		EventType: eventType,
		Session:   session,
	})
}

func (b *SessionLifetimeEventBus) Close() {
	b.pubSub.Close()
}

func (b *SessionLifetimeEventBus) GetHistory() []SessionBuildEvent {
	history := b.pubSub.GetHistory("session")
	sessionEvents := []SessionBuildEvent{}
	for _, rawEvent := range history.GetEntries() {
		sessionEvent, ok := rawEvent.(SessionBuildEvent)
		if ok {
			sessionEvents = append(sessionEvents, sessionEvent)
		}
	}
	return sessionEvents
}
