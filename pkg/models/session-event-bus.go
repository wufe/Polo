package models

import (
	"github.com/wufe/polo/pkg/background/communication"
)

const (
	SessionEventTypeBuildStarted             SessionEventType = "build_started"
	SessionEventTypePreparingFolders         SessionEventType = "preparing_folders"
	SessionEventTypePreparingFoldersFailed   SessionEventType = "preparing_folders_failed"
	SessionEventTypeCommandsExecutionStarted SessionEventType = "commands_execution_started"
	SessionEventTypeCommandsExecutionFailed  SessionEventType = "commands_execution_failed"
	SessionEventTypeWarmupStarted            SessionEventType = "warmup_started"
	SessionEventTypeWarmupFailed             SessionEventType = "warmup_failed"
	SessionEventTypeHealthcheckStarted       SessionEventType = "healthcheck_started"
	SessionEventTypeHealthcheckFailed        SessionEventType = "healthcheck_failed"
	SessionEventTypeStarted                  SessionEventType = "started"
	SessionEventTypeGettingRecycled          SessionEventType = "getting_recycled"
	SessionEventTypeBuildGettingRetried      SessionEventType = "build_getting_retried"
	SessionEventTypeFolderClean              SessionEventType = "folder_clean"
)

type SessionEventType string

func (t SessionEventType) String() string {
	return string(t)
}

type SessionBuildEvent struct {
	EventType SessionEventType
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

func (b *SessionLifetimeEventBus) PublishEvent(eventType SessionEventType, session *Session) {
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
