package node

import (
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/multiformats/go-multiaddr"

	"github.com/makerdao/oracle-suite/pkg/log"
)

func ConnectionLogger() Options {
	return func(n *Node) error {
		cl := &connectionLoggerNotifee{log: n.log}
		n.AddNotifee(cl)
		return nil
	}
}

type connectionLoggerNotifee struct {
	log log.Logger
}

func (n *connectionLoggerNotifee) Listen(network.Network, multiaddr.Multiaddr) {}

func (n *connectionLoggerNotifee) ListenClose(network.Network, multiaddr.Multiaddr) {}

func (n *connectionLoggerNotifee) Connected(network network.Network, conn network.Conn) {
	n.log.
		WithFields(log.Fields{"ip": conn.LocalMultiaddr().String()}).
		Debug("Connected to a host")
}

func (n *connectionLoggerNotifee) Disconnected(network network.Network, conn network.Conn) {
	n.log.
		WithFields(log.Fields{"ip": conn.LocalMultiaddr().String()}).
		Debug("Disconnected from a host")
}

func (n *connectionLoggerNotifee) OpenedStream(network.Network, network.Stream) {}

func (n *connectionLoggerNotifee) ClosedStream(network.Network, network.Stream) {}
