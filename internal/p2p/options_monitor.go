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
