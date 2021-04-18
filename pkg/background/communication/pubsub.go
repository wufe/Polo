package communication

import (
	"github.com/wufe/polo/pkg/utils"
)

// PubSub represents a structure capable of performing publish and subscribe
type PubSub struct {
	mutex         utils.RWLocker
	history       map[string]*PubSubHistory
	subscriptions map[string][]chan interface{}
	closed        bool
	mutexBuilder  utils.MutexBuilder
}

func newPubSub(mutexBuilder utils.MutexBuilder) *PubSub {
	return &PubSub{
		mutex:         mutexBuilder(),
		history:       make(map[string]*PubSubHistory),
		subscriptions: make(map[string][]chan interface{}),
		mutexBuilder:  mutexBuilder,
	}
}

// Publish an event by its topic
func (ps *PubSub) Publish(topic string, payload interface{}) {
	ps.mutex.RLock()
	defer ps.mutex.RUnlock()

	if ps.closed {
		return
	}

	_, exists := ps.history[topic]
	if !exists {
		ps.mutex.RUnlock()
		ps.createHistory(topic)
		ps.mutex.RLock()
	}

	for _, ch := range ps.subscriptions[topic] {
		go func(ch chan interface{}) {
			ch <- payload
		}(ch)
	}
}

func (ps *PubSub) createHistory(topic string) *PubSubHistory {
	ps.mutex.Lock()
	history := &PubSubHistory{
		mutex: ps.mutexBuilder(),
	}
	ps.history[topic] = history

	// History filling subscription
	ch := ps.subscribeInternal(topic)
	go func(history *PubSubHistory) {
		for {
			ev, ok := <-ch
			if !ok {
				return
			}
			history.mutex.Lock()
			history.entries = append(history.entries, ev)
			history.mutex.Unlock()
		}
	}(history)

	ps.mutex.Unlock()
	return history
}

// Subscribe to a topic
func (ps *PubSub) Subscribe(topic string) <-chan interface{} {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()

	return ps.subscribeInternal(topic)
}

func (ps *PubSub) subscribeInternal(topic string) <-chan interface{} {
	ch := make(chan interface{})
	ps.subscriptions[topic] = append(ps.subscriptions[""], ch)
	return ch
}

// Close the pubsub subscriptions
func (ps *PubSub) Close() {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()

	if !ps.closed {
		ps.closed = true
		for _, subscription := range ps.subscriptions {
			for _, ch := range subscription {
				close(ch)
			}
		}
	}
}

// GetHistory retrieves the history of the events for a given topic
func (ps *PubSub) GetHistory(topic string) *PubSubHistory {
	ps.mutex.RLock()
	defer ps.mutex.RUnlock()
	history, exists := ps.history[topic]
	if !exists {
		ps.mutex.RUnlock()
		history = ps.createHistory(topic)
		ps.mutex.RLock()
	}
	return history
}

// PubSubHistory is the object containing history events
type PubSubHistory struct {
	mutex   utils.RWLocker
	entries []interface{}
}

// GetEntries retrieves the history events from a PubSubHistory object
func (h *PubSubHistory) GetEntries() []interface{} {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return h.entries
}
