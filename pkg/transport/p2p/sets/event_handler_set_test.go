package sets

import (
	"testing"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/stretchr/testify/assert"
)

func TestEventHandlerSet_Handle(t *testing.T) {
	ehs := NewEventHandlerSet()

	pe := pubsub.PeerEvent{
		Type: 1,
		Peer: "a",
	}

	// All event handlers should be invoked:
	calls := 0
	ehs.Add(EventHandlerFunc(func(topic string, event pubsub.PeerEvent) {
		assert.Equal(t, "foo", topic)
		assert.Equal(t, pe, event)
		calls++
	}))
	ehs.Add(EventHandlerFunc(func(topic string, event pubsub.PeerEvent) {
		assert.Equal(t, "foo", topic)
		assert.Equal(t, pe, event)
		calls++
	}))
	ehs.Handle("foo", pe)
	assert.Equal(t, 2, calls)
}
