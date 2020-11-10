package logger

import (
	pubsub "github.com/libp2p/go-libp2p-pubsub"

	"github.com/makerdao/gofer/internal/log"
	"github.com/makerdao/gofer/internal/transport"
)

type messageHandler struct {
	log log.Logger
}

func (m *messageHandler) Published(topic string, raw []byte, message transport.Message) {
	m.log.
		WithFields(log.Fields{"topic": topic, "message": string(raw)}).
		Debug("Published new message")
}

func (m *messageHandler) Received(topic string, raw *pubsub.Message, message transport.Message) {
	m.log.
		WithFields(log.Fields{"topic": topic, "message": string(raw.Data), "peerID": raw.ReceivedFrom}).
		Debug("Received new message")
}
