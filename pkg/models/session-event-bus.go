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
	pubSub *communication.PubSub
}

func NewSessionBuildEventBus(pubSubBuilder *communication.PubSubBuilder) *SessionLifetimeEventBus {
	pubSub := pubSubBuilder.Build()
	eventBus := &SessionLifetimeEventBus{
		pubSub: pubSub,
	}
	return eventBus
}

func (b *SessionLifetimeEventBus) GetChan() <-chan SessionBuildEvent {
	sourceCh, history := b.pubSub.Subscribe("session")
	destCh := make(chan SessionBuildEvent)
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
			sessionEv, ok := ev.(SessionBuildEvent)
			if !ok {
				continue
			}
			destCh <- sessionEv
		}
	}()
	return destCh
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

func (b *SessionLifetimeEventBus) convertHistoryEntries(entries []interface{}) []SessionBuildEvent {
	sessionEvents := []SessionBuildEvent{}
	for _, rawEvent := range entries {
		sessionEvent, ok := rawEvent.(SessionBuildEvent)
		if ok {
			sessionEvents = append(sessionEvents, sessionEvent)
		}
	}
	return sessionEvents
}
