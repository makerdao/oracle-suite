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

package node

import (
	pubsub "github.com/libp2p/go-libp2p-pubsub"

	"github.com/makerdao/oracle-suite/pkg/log"
	"github.com/makerdao/oracle-suite/pkg/transport/p2p/node/sets"
)

// PeerLogger logs all peers handled by libp2p's pubsub system.
func PeerLogger() Options {
	return func(n *Node) error {
		n.AddPubSubEventHandler(sets.PubSubEventHandlerFunc(func(topic string, event pubsub.PeerEvent) {
			addrs := n.Host().Peerstore().PeerInfo(event.Peer).Addrs
			switch event.Type {
			case pubsub.PeerJoin:
				n.log.
					WithFields(log.Fields{
						"peerID": event.Peer.String(),
						"topic":  topic,
						"addrs":  addrs,
					}).
					Debug("Connected to a peer")
			case pubsub.PeerLeave:
				n.log.
					WithFields(log.Fields{
						"peerID": event.Peer.String(),
						"topic":  topic,
						"addrs":  addrs,
					}).
					Debug("Disconnected from a peer")
			}
		}))
		return nil
	}
}