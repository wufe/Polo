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
	options       PubSubOptions
}

type PubSubOptions struct {
	Buffer int
}

func newPubSub(mutexBuilder utils.MutexBuilder, options PubSubOptions) *PubSub {
	return &PubSub{
		mutex:         mutexBuilder(),
		history:       make(map[string]*PubSubHistory),
		subscriptions: make(map[string][]chan interface{}),
		mutexBuilder:  mutexBuilder,
		options:       options,
	}
}

// Publish an event by its topic
func (ps *PubSub) Publish(topic string, payload interface{}) {
	ps.mutex.RLock()
	defer ps.mutex.RUnlock()

	if ps.closed {
		return
	}

	history, exists := ps.history[topic]
	if !exists {
		ps.mutex.RUnlock()
		ps.mutex.Lock()
		history = ps.createHistoryInternal(topic)
		ps.mutex.Unlock()
		ps.mutex.RLock()
	}

	history.AddEntry(payload)

	for _, ch := range ps.subscriptions[topic] {
		ch <- payload
	}
}

func (ps *PubSub) createHistoryInternal(topic string) *PubSubHistory {
	history := &PubSubHistory{
		mutex: ps.mutexBuilder(),
	}
	ps.history[topic] = history

	return history
}

// Subscribe to a topic
func (ps *PubSub) Subscribe(topic string) (<-chan interface{}, []interface{}) {
	ps.mutex.Lock()
	subscriptionChan := ps.subscribeInternal(topic)
	defer ps.mutex.Unlock()

	history := ps.getHistoryInternal(topic)

	return subscriptionChan, history.GetEntries()
}

func (ps *PubSub) subscribeInternal(topic string) <-chan interface{} {
	ch := make(chan interface{}, ps.options.Buffer)
	ps.subscriptions[topic] = append(ps.subscriptions[topic], ch)
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

func (ps *PubSub) getHistoryInternal(topic string) *PubSubHistory {
	history, exists := ps.history[topic]
	if !exists {
		history = ps.createHistoryInternal(topic)
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

func (h *PubSubHistory) AddEntry(entry interface{}) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	h.entries = append(h.entries, entry)
}
