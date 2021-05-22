package models

import (
	"github.com/asaskevich/EventBus"
	"github.com/wufe/polo/pkg/utils"
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
	SessionEventTypeHealthcheckSucceded      SessionEventType = "healthcheck_succeeded"
	SessionEventTypeStarted                  SessionEventType = "started"
	SessionEventTypeBuildGettingRetried      SessionEventType = "build_getting_retried"
	SessionEventTypeFolderClean              SessionEventType = "folder_clean"
	SessionEventTypeCleanCommandExecution    SessionEventType = "clean_command_execution"
	SessionEventTypeSessionAvailable         SessionEventType = "session_available"
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
	utils.RWLocker
	bus     EventBus.Bus
	ch      chan SessionBuildEvent
	history []SessionBuildEvent
}

func NewSessionBuildEventBus(mutexBuilder utils.MutexBuilder) *SessionLifetimeEventBus {
	bus := EventBus.New()
	eventBus := &SessionLifetimeEventBus{
		RWLocker: mutexBuilder(),
		bus:      bus,
		ch:       make(chan SessionBuildEvent, eventsBuffer),
	}
	eventBus.start()
	return eventBus
}

func (b *SessionLifetimeEventBus) start() {
	history := []SessionBuildEvent{}
	b.history = history

	eventsCount := 0

	b.bus.SubscribeAsync("session", func(ev interface{}) {
		if sessionEv, ok := ev.(SessionBuildEvent); ok {
			b.Lock()
			defer b.Unlock()
			history = append(history, sessionEv)
			b.ch <- sessionEv
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
	}, true)
}

func (b *SessionLifetimeEventBus) GetChan() <-chan SessionBuildEvent {
	return b.ch
}

func (b *SessionLifetimeEventBus) PublishEvent(eventType SessionEventType, session *Session) {
	b.bus.Publish("session", SessionBuildEvent{
		EventType: eventType,
		Session:   session,
	})
}

func (b *SessionLifetimeEventBus) Close() {}

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
