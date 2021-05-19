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
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	discovery "github.com/libp2p/go-libp2p-discovery"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/multiformats/go-multiaddr"

	"github.com/makerdao/oracle-suite/pkg/log"
	"github.com/makerdao/oracle-suite/pkg/transport/p2p/node/sets"
)

type Options func(n *Node) error

// ListenAddrs configures node to listen on the given addresses.
func ListenAddrs(addrs []multiaddr.Multiaddr) Options {
	return func(n *Node) error {
		n.listenAddrs = addrs
		return nil
	}
}

// PeerPrivKey configures node to use given key as its identity.
func PeerPrivKey(pk crypto.PrivKey) Options {
	return func(n *Node) error {
		n.peerPrivKey = pk
		return nil
	}
}

// MessagePrivKey configures node to use given key to sign messages.
func MessagePrivKey(pk crypto.PrivKey) Options {
	return func(n *Node) error {
		pid, err := peer.IDFromPublicKey(pk.GetPublic())
		if err != nil {
			return err
		}
		n.messagePrivKey = pk
		n.messageAuthorPID = pid
		return nil
	}
}

// Logger configures node to use given logger instance.
func Logger(logger log.Logger) Options {
	return func(n *Node) error {
		n.log = logger
		return nil
	}
}

// Bootstrap configures node to use given list of addresses as bootstrap nodes.
func Bootstrap(addrs []multiaddr.Multiaddr) Options {
	return func(n *Node) error {
		n.AddNodeEventHandler(sets.NodeEventHandlerFunc(func(event sets.NodeEventType) {
			if event != sets.NodeStarted {
				return
			}
			for _, maddr := range addrs {
				err := n.Connect(maddr)
				if err != nil {
					n.log.
						WithFields(log.Fields{"addr": maddr.String()}).
						WithError(err).
						Warn("Unable to connect to the bootstrap peer")
				}
			}
		}))
		return nil
	}
}

// DHT configures node to use kad-dht for a node discovery.
func DHT(rendezvousString string) Options {
	return func(n *Node) error {
		var err error
		var kaddht *dht.IpfsDHT

		n.AddNodeEventHandler(sets.NodeEventHandlerFunc(func(event sets.NodeEventType) {
			switch event {
			case sets.NodeStarted:
				// Initialize DHT:
				kaddht, err = dht.New(n.ctx, n.host)
				if err != nil {
					n.log.
						WithError(err).
						Error("Unable to initialize DHT")
				}
				// Use a rendezvous point to announce our location:
				routingDiscovery := discovery.NewRoutingDiscovery(kaddht)
				discovery.Advertise(n.ctx, routingDiscovery, rendezvousString)
			case sets.NodeStopping:
				// Close DHT:
				err = kaddht.Close()
				if err != nil {
					n.log.
						WithError(err).
						Error("Unable to close DHT")
				}
			}
		}))
		return nil
	}
}
