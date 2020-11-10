package sets

import (
	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

// EventHandler can ba implemented by type that supports handling the PubSub
// system events.
type EventHandler interface {
	// Handle is called on a new event.
	Handle(topic string, event pubsub.PeerEvent)
}

// EventHandlerSet stores multiple instances of the EventHandler interface.
type EventHandlerSet struct {
	eventHandler []EventHandler
}

// NewEventHandlerSet creates new instance of the EventHandlerSet.
func NewEventHandlerSet() *EventHandlerSet {
	return &EventHandlerSet{}
}

// Add adds new EventHandler to the set.
func (n *EventHandlerSet) Add(eventHandler ...EventHandler) {
	n.eventHandler = append(n.eventHandler, eventHandler...)
}

// Handle invokes all registered handlers for given topic.
func (n *EventHandlerSet) Handle(topic string, event pubsub.PeerEvent) {
	for _, eventHandler := range n.eventHandler {
		eventHandler.Handle(topic, event)
	}
}
