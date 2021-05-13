package node

import (
	"github.com/libp2p/go-libp2p-core/peerstore"
	pubsub "github.com/libp2p/go-libp2p-pubsub"

	"github.com/makerdao/oracle-suite/pkg/log"
)

func PeerLogger() Options {
	return func(n *Node) error {
		eh := &peerLoggerEventHandler{peerstore: n.Host().Peerstore(), log: n.log}
		n.AddEventHandler(eh)
		return nil
	}
}

type peerLoggerEventHandler struct {
	peerstore peerstore.Peerstore
	log       log.Logger
}

func (e *peerLoggerEventHandler) Handle(topic string, event pubsub.PeerEvent) {
	addrs := e.peerstore.PeerInfo(event.Peer).Addrs

	switch event.Type {
	case pubsub.PeerJoin:
		e.log.
			WithFields(log.Fields{
				"peerID": event.Peer.String(),
				"topic":  topic,
				"addrs":  addrs,
			}).
			Debug("Connected to a peer")
	case pubsub.PeerLeave:
		e.log.
			WithFields(log.Fields{
				"peerID": event.Peer.String(),
				"topic":  topic,
				"addrs":  addrs,
			}).
			Debug("Disconnected from a peer")
	}
}
