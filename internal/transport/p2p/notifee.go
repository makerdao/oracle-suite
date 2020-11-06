package p2p

import (
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/multiformats/go-multiaddr"

	"github.com/makerdao/gofer/internal/log"
)

type notifee struct {
	log log.Logger
}

func newNotifee(l log.Logger) *notifee {
	return &notifee{
		log: l,
	}
}

func (n *notifee) Listen(network.Network, multiaddr.Multiaddr) {}

func (n *notifee) ListenClose(network.Network, multiaddr.Multiaddr) {}

func (n *notifee) Connected(network network.Network, conn network.Conn) {
	n.log.
		WithFields(log.Fields{"ip": conn.LocalMultiaddr().String()}).
		Debug("Connected to peer")
}

func (n *notifee) Disconnected(network network.Network, conn network.Conn) {
	n.log.
		WithFields(log.Fields{"ip": conn.LocalMultiaddr().String()}).
		Debug("Disconnected from peer")
}

func (n *notifee) OpenedStream(network.Network, network.Stream) {}

func (n *notifee) ClosedStream(network.Network, network.Stream) {}
