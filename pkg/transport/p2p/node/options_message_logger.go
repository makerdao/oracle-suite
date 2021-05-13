package node

import (
	"github.com/libp2p/go-libp2p-core/peerstore"
	pubsub "github.com/libp2p/go-libp2p-pubsub"

	"github.com/makerdao/oracle-suite/pkg/log"
	"github.com/makerdao/oracle-suite/pkg/transport"
)

func MessageLogger() Options {
	return func(n *Node) error {
		mlh := &messageLoggerHandler{log: n.log}
		n.AddMessageHandler(mlh)
		return nil
	}
}

type messageLoggerHandler struct {
	peerStore peerstore.Peerstore
	log       log.Logger
}

func (m *messageLoggerHandler) Published(topic string, raw []byte, message transport.Message) {
	m.log.
		WithFields(log.Fields{"topic": topic, "message": string(raw)}).
		Debug("Published a new message")
}

func (m *messageLoggerHandler) Received(topic string, raw *pubsub.Message, message transport.Message) {
	addrs := m.peerStore.PeerInfo(raw.ReceivedFrom).Addrs
	m.log.
		WithFields(log.Fields{
			"topic":              topic,
			"message":            string(raw.Data),
			"peerID":             raw.GetFrom().String(),
			"receivedFromPeerID": raw.ReceivedFrom.String(),
			"receivedFromAddrs":  addrs,
		}).
		Debug("Received a new message")
}
