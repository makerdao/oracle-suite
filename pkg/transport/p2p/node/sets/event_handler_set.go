//  Copyright (C) 2020 Maker Ecosystem Growth Holdings, INC.
//
//  This program is free software: you can redistribute it and/or modify
//  it under the terms of the GNU Affero General Public License as
//  published by the Free Software Foundation, either version 3 of the
//  License, or (at your option) any later version.
//
//  This program is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU Affero General Public License for more details.
//
//  You should have received a copy of the GNU Affero General Public License
//  along with this program.  If not, see <http://www.gnu.org/licenses/>.

package sets

import (
	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

// EventHandlerFunc is a adapter for the EventHandler interface.
type EventHandlerFunc func(topic string, event pubsub.PeerEvent)

// Handle calls f(topic, event).
func (f EventHandlerFunc) Handle(topic string, event pubsub.PeerEvent) {
	f(topic, event)
}

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

var _ EventHandler = (*EventHandlerSet)(nil)
