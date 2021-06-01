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
	"github.com/multiformats/go-multiaddr"

	"github.com/makerdao/oracle-suite/internal/p2p"
	"github.com/makerdao/oracle-suite/pkg/ethereum"
	"github.com/makerdao/oracle-suite/pkg/log"
	"github.com/makerdao/oracle-suite/pkg/transport"
)

const LoggerTag = "P2P"
const userAgentString = "spire/v0.0-dev"

// defaultListenAddrs is a list of default multiaddresses on which node will
// be listening on.
var defaultListenAddrs = []string{"/ip4/0.0.0.0/tcp/0"}

// P2P is a little wrapper for the Node that implements the Transport
// interface.
type P2P struct {
	node *p2p.Node
}

type Config struct {
	Context context.Context
	Logger  log.Logger

	// PeerPrivKey is a key used for peer identity. If empty, then random key
	// is used.
	PeerPrivKey crypto.PrivKey
	// MessagePrivKey is a key used to sign messages. If empty, then message
	// are signed with the same key which is used for peer identity.
	MessagePrivKey crypto.PrivKey
	// ListenAddrs is a list of multiaddresses on which this node will be
	// listening on. If empty, the localhost, and a random port will be used.
	ListenAddrs []string
	// BootstrapAddrs is a list multiaddresses of initial peers to connect to.
	BootstrapAddrs []string
	// BlockedAddrs is a list of multiaddresses to which connection will be
	// blocked. If an address on that list contains an IP and a peer ID, both
	// will be blocked separately.
	BlockedAddrs []string
	// FeedersAddrs is a list of price feeders. Only feeders can create new
	// messages in the network.
	FeedersAddrs []ethereum.Address
	// Discovery indicates whenever Discovery should be enabled.
	Discovery bool
	// Signer used to verify price messages.
	Signer ethereum.Signer
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

	logger := cfg.Logger.WithField("tag", LoggerTag)
	opts := []p2p.Options{
		p2p.UserAgent(userAgentString),
		p2p.ListenAddrs(listenAddrs),
		p2p.Denylist(blockedAddrs),
		p2p.Logger(logger),
		p2p.ConnectionLogger(),
		p2p.MessageLogger(),
		p2p.PeerLogger(),
		oracle(cfg.FeedersAddrs, cfg.Signer, logger),
	}

	if cfg.PeerPrivKey != nil {
		opts = append(opts, p2p.PeerPrivKey(cfg.PeerPrivKey))
	}

	if cfg.MessagePrivKey != nil {
		opts = append(opts, p2p.MessagePrivKey(cfg.MessagePrivKey))
	}

	if cfg.Discovery {
		opts = append(opts, p2p.Discovery(bootstrapAddrs))
	} else {
		opts = append(opts, p2p.Bootstrap(bootstrapAddrs))
	}

	n, err := p2p.NewNode(cfg.Context, opts...)
	if err != nil {
		return nil, err
	}

	p := &P2P{node: n}
	err = p.node.Start()
	if err != nil {
		return nil, err
	}

	return p, nil
}

// Subscribe implements the transport.Transport interface.
func (p *P2P) Subscribe(topic string, typ transport.Message) error {
	return p.node.Subscribe(topic, typ)
}

// Unsubscribe implements the transport.Transport interface.
func (p *P2P) Unsubscribe(topic string) error {
	return p.node.Unsubscribe(topic)
}

// Broadcast implements the transport.Transport interface.
func (p *P2P) Broadcast(topic string, message transport.Message) error {
	sub, err := p.node.Subscription(topic)
	if err != nil {
		return err
	}
	return sub.Publish(message)
}

// WaitFor implements the transport.Transport interface.
func (p *P2P) WaitFor(topic string) chan transport.Status {
	sub, err := p.node.Subscription(topic)
	if err != nil {
		return nil
	}
	return sub.Next()
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
