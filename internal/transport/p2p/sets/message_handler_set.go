package sets

import (
	pubsub "github.com/libp2p/go-libp2p-pubsub"

	"github.com/makerdao/gofer/internal/transport"
)

// MessageHandler can ba implemented by type that supports handling the PubSub
// system messages.
type MessageHandler interface {
	// Handle is called when new message is published.
	Published(topic string, raw []byte, message transport.Message)
	// Handle is called when new message is received.
	Received(topic string, raw *pubsub.Message, message transport.Message)
}

// MessageHandlerSet stores multiple instances of the MessageHandler interface.
type MessageHandlerSet struct {
	messageHandler []MessageHandler
}

// NewMessageHandlerSet creates new instance of the MessageHandlerSet.
func NewMessageHandlerSet() *MessageHandlerSet {
	return &MessageHandlerSet{}
}

// Add adds new MessageHandler to the set.
func (n *MessageHandlerSet) Add(messageHandler ...MessageHandler) {
	n.messageHandler = append(n.messageHandler, messageHandler...)
}

// Handle invokes all registered handlers for given topic.
func (n *MessageHandlerSet) Published(topic string, raw []byte, message transport.Message) {
	if raw == nil {
		return
	}
	for _, messageHandler := range n.messageHandler {
		messageHandler.Published(topic, raw, message)
	}
}

// Handle invokes all registered handlers for given topic.
func (n *MessageHandlerSet) Received(topic string, raw *pubsub.Message, message transport.Message) {
	if raw == nil {
		return
	}
	for _, messageHandler := range n.messageHandler {
		messageHandler.Received(topic, raw, message)
	}
}
