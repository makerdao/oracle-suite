package logger

import (
	pubsub "github.com/libp2p/go-libp2p-pubsub"

	"github.com/makerdao/gofer/internal/log"
)

type eventHandler struct {
	log log.Logger
}

func (e eventHandler) Handle(topic string, event pubsub.PeerEvent) {
	switch event.Type {
	case pubsub.PeerJoin:
		e.log.
			WithFields(log.Fields{"id": event.Peer.Pretty(), "topic": topic}).
			Debug("Connected to peer")
	case pubsub.PeerLeave:
		e.log.
			WithFields(log.Fields{"id": event.Peer.Pretty(), "topic": topic}).
			Debug("Disconnected from peer")
	}
}
