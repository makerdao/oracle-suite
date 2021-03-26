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
	"context"

	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"

	"github.com/makerdao/oracle-suite/pkg/log"
	"github.com/makerdao/oracle-suite/pkg/transport"
)

const LoggerTag = "P2P"

// defaultListenAddrs is a list of default multiaddresses on which node will
// be listening on.
var defaultListenAddrs = []string{"/ip4/0.0.0.0/tcp/0"}

type P2P struct {
	node *Node
}

type Config struct {
	Context context.Context
	Logger  log.Logger

	// PrivateKey is a key used to sign and verify messages in the network.
	PrivateKey crypto.PrivKey
	// ListenAddrs is a list of multiaddresses on which this node will be
	// listening on. If empty, the localhost, and a random port will be used.
	ListenAddrs []string
	// BootstrapAddrs is a list multiaddresses of initial peers to connect to.
	BootstrapAddrs []string
	// BlockedAddrs is a list of multiaddresses to which connection will be
	// blocked. If an address on that list contains an IP and a peer ID, both
	// will be blocked separately.
	BlockedAddrs []string
	// AllowedPeers is a list of peer IDs which are allowed to publish messages
	// to the network. Messages from peers outside this list will be ignored
	// and not relayed. If empty, all messages will be accepted.
	AllowedPeers []string
}

// New returns a new instance of a transport, implemented with
// the libp2p library.
func New(cfg Config) (*P2P, error) {
	var err error

	if len(cfg.ListenAddrs) == 0 {
		cfg.ListenAddrs = defaultListenAddrs
	}

	listenAddrs, err := strsToMaddrs(cfg.ListenAddrs)
	if err != nil {
		return nil, err
	}
	bootstrapAddrs, err := strsToMaddrs(cfg.BootstrapAddrs)
	if err != nil {
		return nil, err
	}
	blockedAddrs, err := strsToMaddrs(cfg.BlockedAddrs)
	if err != nil {
		return nil, err
	}
	allowedPeers, err := strsToPeerIDs(cfg.AllowedPeers)
	if err != nil {
		return nil, err
	}

	p := &P2P{
		node: NewNode(NodeConfig{
			Context:        cfg.Context,
			Logger:         cfg.Logger.WithField("tag", LoggerTag),
			ListenAddrs:    listenAddrs,
			BootstrapAddrs: bootstrapAddrs,
			BlockedAddrs:   blockedAddrs,
			AllowedPeers:   allowedPeers,
			PrivateKey:     cfg.PrivateKey,
		}),
	}

	err = p.node.Start()
	if err != nil {
		return nil, err
	}

	return p, nil
}

// Subscribe implements the transport.Transport interface.
func (p *P2P) Subscribe(topic string) error {
	return p.node.Subscribe(topic)
}

// Unsubscribe implements the transport.Transport interface.
func (p *P2P) Unsubscribe(topic string) error {
	return p.node.Unsubscribe(topic)
}

// Broadcast implements the transport.Transport interface.
func (p *P2P) Broadcast(topic string, message transport.Message) error {
	sub, err := p.node.subscription(topic)
	if err != nil {
		return err
	}
	return sub.Publish(message)
}

// WaitFor implements the transport.Transport interface.
func (p *P2P) WaitFor(topic string, message transport.Message) chan transport.Status {
	sub, err := p.node.subscription(topic)
	if err != nil {
		return nil
	}
	return sub.Next(message)
}

// Close implements the transport.Transport interface.
func (p *P2P) Close() error {
	return p.node.Stop()
}

// strsToMaddrs converts multiaddresses given as strings to a
// list of multiaddr.Multiaddr.
func strsToMaddrs(addrs []string) ([]multiaddr.Multiaddr, error) {
	var maddrs []multiaddr.Multiaddr
	for _, addrstr := range addrs {
		maddr, err := multiaddr.NewMultiaddr(addrstr)
		if err != nil {
			return nil, err
		}
		maddrs = append(maddrs, maddr)
	}
	return maddrs, nil
}

// strsToPeerIDs converts peer IDs given as strings to a
// list of peer.ID.
func strsToPeerIDs(ids []string) ([]peer.ID, error) {
	var pIDs []peer.ID
	for _, s := range ids {
		pID, err := peer.Decode(s)
		if err != nil {
			return nil, err
		}
		pIDs = append(pIDs, pID)
	}
	return pIDs, nil
}
