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

	"github.com/libp2p/go-libp2p"
	relay "github.com/libp2p/go-libp2p-circuit"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	discovery "github.com/libp2p/go-libp2p-discovery"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/multiformats/go-multiaddr"

	"github.com/makerdao/oracle-suite/internal/p2p/sets"
	"github.com/makerdao/oracle-suite/pkg/log"
)

type Options func(n *Node) error

// ListenAddrs configures node to listen on the given addresses.
func ListenAddrs(addrs []multiaddr.Multiaddr) Options {
	return func(n *Node) error {
		n.hostOpts = append(n.hostOpts, libp2p.ListenAddrs(addrs...))
		return nil
	}
}

// PeerPrivKey configures node to use given key as its identity.
func PeerPrivKey(sk crypto.PrivKey) Options {
	return func(n *Node) error {
		n.hostOpts = append(n.hostOpts, libp2p.Identity(sk))
		return nil
	}
}

// MessagePrivKey configures node to use given key to sign messages.
func MessagePrivKey(sk crypto.PrivKey) Options {
	return func(n *Node) error {
		pid, err := peer.IDFromPublicKey(sk.GetPublic())
		if err != nil {
			return err
		}

		// It's necessary to add this key to the peerstore,
		// otherwise it'll be impossible to use it to sign messages:
		err = n.Peerstore().AddPrivKey(pid, sk)
		if err != nil {
			return err
		}

		n.pubsubOpts = append(n.pubsubOpts, pubsub.WithMessageAuthor(pid))
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

// UserAgent sets the libp2p user-agent sent along with the identify protocol.
func UserAgent(userAgent string) Options {
	return func(n *Node) error {
		n.hostOpts = append(n.hostOpts, libp2p.UserAgent(userAgent))
		return nil
	}
}

// CircuitRelay configures node to use circuit relay.
func CircuitRelay(relayAddrs []multiaddr.Multiaddr) Options {
	return func(n *Node) error {
		addrs, err := peer.AddrInfosFromP2pAddrs(relayAddrs...)
		if err != nil {
			return err
		}

		n.hostOpts = append(n.hostOpts, libp2p.EnableAutoRelay())
		if len(addrs) > 0 {
			n.hostOpts = append(
				n.hostOpts,
				libp2p.EnableRelay(),
				libp2p.StaticRelays(addrs),
			)
		} else {
			n.hostOpts = append(
				n.hostOpts,
				libp2p.EnableRelay(relay.OptHop),
			)
		}

		return nil
	}
}

// Bootstrap configures node to use given list of addresses as bootstrap nodes.
func Bootstrap(addrs []multiaddr.Multiaddr) Options {
	return func(n *Node) error {
		var addrInfos []*peer.AddrInfo
		for _, maddr := range addrs {
			ai, err := peer.AddrInfoFromP2pAddr(maddr)
			if err != nil {
				return err
			}
			addrInfos = append(addrInfos, ai)
		}

		n.AddNodeEventHandler(sets.NodeEventHandlerFunc(func(event sets.NodeEventType) {
			if event != sets.NodeStarted {
				return
			}
			connect := func() {
				for _, addrInfo := range addrInfos {
					for _, maddr := range addrInfo.Addrs {
						if n.host.Network().Connectedness(addrInfo.ID) != network.NotConnected {
							continue
						}
						err := n.Connect(maddr)
						if err != nil {
							n.log.
								WithFields(log.Fields{"addr": addrInfo.String()}).
								WithError(err).
								Warn("Unable to connect to the bootstrap peer")
						}
					}
				}
			}
			go func() {
				t := time.NewTimer(2 * time.Minute)
				connect()
				select {
				case <-n.ctx.Done():
					t.Stop()
					return
				case <-t.C:
					connect()
				}
			}()
		}))
		return nil
	}
}

// Discovery configures node to use kad-dht for node discovery.
func Discovery(bootstrapAddrs []multiaddr.Multiaddr) Options {
	return func(n *Node) error {
		var err error
		var kadDHT *dht.IpfsDHT

		addrs, err := peer.AddrInfosFromP2pAddrs(bootstrapAddrs...)
		if err != nil {
			return err
		}

		n.AddNodeEventHandler(sets.NodeEventHandlerFunc(func(event sets.NodeEventType) {
			switch event {
			case sets.NodeHostStarted:
				kadDHT, err = dht.New(n.ctx, n.host, dht.BootstrapPeers(addrs...), dht.Mode(dht.ModeServer))
				if err != nil {
					n.log.
						WithError(err).
						Error("Unable to initialize KAD-DHT")
					return
				}
				if err = kadDHT.Bootstrap(n.ctx); err != nil {
					n.log.
						WithError(err).
						Error("Unable to bootstrap KAD-DHT")
					return
				}
				n.pubsubOpts = append(n.pubsubOpts, pubsub.WithDiscovery(discovery.NewRoutingDiscovery(kadDHT)))
			case sets.NodeStopping:
				if kadDHT == nil {
					return
				}
				err = kadDHT.Close()
				if err != nil {
					n.log.
						WithError(err).
						Error("Unable to close KAD-DHT")
				}
			}
		}))
		return nil
	}
}
