package p2p

import (
	"time"

	"github.com/libp2p/go-libp2p-core/network"
	"github.com/multiformats/go-multiaddr"

	"github.com/makerdao/oracle-suite/internal/p2p/sets"
)

type monitorNotifee struct {
	notifeeCh chan struct{}
}

// Listen implements the network.Notifiee interface.
func (n *monitorNotifee) Listen(network.Network, multiaddr.Multiaddr) {}

// ListenClose implements the network.Notifiee interface.
func (n *monitorNotifee) ListenClose(network.Network, multiaddr.Multiaddr) {}

// Connected implements the network.Notifiee interface.
func (n *monitorNotifee) Connected(_ network.Network, conn network.Conn) {
	n.notifeeCh <- struct{}{}
}

// Disconnected implements the network.Notifiee interface.
func (n *monitorNotifee) Disconnected(_ network.Network, conn network.Conn) {
	n.notifeeCh <- struct{}{}
}

// OpenedStream implements the network.Notifiee interface.
func (n *monitorNotifee) OpenedStream(network.Network, network.Stream) {}

// ClosedStream implements the network.Notifiee interface.
func (n *monitorNotifee) ClosedStream(network.Network, network.Stream) {}

func Monitor() Options {
	log := func(n *Node) {
		n.log.
			WithField("peerCount", len(n.host.Network().Peers())).
			Info("Connected peers")
	}
	return func(n *Node) error {
		notifeeCh := make(chan struct{})
		n.AddNotifee(&monitorNotifee{notifeeCh: notifeeCh})
		n.AddNodeEventHandler(sets.NodeEventHandlerFunc(func(event interface{}) {
			if _, ok := event.(sets.NodeStartedEvent); ok {
				go func() {
					t := time.NewTicker(time.Minute)
					for {
						select {
						case <-notifeeCh:
							log(n)
						case <-t.C:
							log(n)
						case <-n.ctx.Done():
							t.Stop()
							return
						}
					}
				}()
			}
		}))
		return nil
	}
}
