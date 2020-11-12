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
	"fmt"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"

	"github.com/makerdao/gofer/internal/ethereum"
	"github.com/makerdao/gofer/internal/log"
	"github.com/makerdao/gofer/internal/transport"
	"github.com/makerdao/gofer/internal/transport/p2p/allowlist"
	"github.com/makerdao/gofer/internal/transport/p2p/banner"
	"github.com/makerdao/gofer/internal/transport/p2p/ethkey"
	"github.com/makerdao/gofer/internal/transport/p2p/logger"
)

const LoggerTag = "P2P"

// defaultListenAddrs is a list of default multiaddresses on which node will
// be listening on.
var defaultListenAddrs = []string{"/ip4/0.0.0.0/tcp/0"}

type P2P struct {
	log       log.Logger
	node      *Node
	allowlist *allowlist.Allowlist
	banner    *banner.Banner
}

type Config struct {
	Context context.Context
	Logger  log.Logger

	// Wallet is used to sign and verify messages in the network.
	Wallet ethereum.Account
	// ListenAddrs is a list of multiaddresses on which node will be
	// listening on. If empty, localhost and random port will be used.
	ListenAddrs []string
	// BootstrapAddrs is a list multiaddresses of initial peers to connect to.
	BootstrapAddrs []string
	// BannedAddrs is a list of multiaddresses to which connection will be
	// blocked. If an address on that list contains an IP and a peer ID, both
	// will be blocked separately.
	BannedAddrs []string
	// AllowedPeers is a list of peer IDs which are allowed to publish messages
	// to the network. Messages from peers outside this list will be ignored
	// and not relayed. If empty, all messages will be accepted.
	AllowedPeers []string
}

// NewP2P returns new instance of transport implemented using the libp2p
// library.
func NewP2P(config Config) (*P2P, error) {
	var err error

	if len(config.ListenAddrs) == 0 {
		config.ListenAddrs = defaultListenAddrs
	}
	listenAddrs, err := addrsToMaddrs(config.ListenAddrs)
	if err != nil {
		return nil, err
	}

	p := &P2P{}
	p.log = log.WrapLogger(config.Logger, log.Fields{"tag": LoggerTag})
	p.node = NewNode(NodeConfig{
		Context:     config.Context,
		Logger:      config.Logger,
		ListenAddrs: listenAddrs,
		PrivateKey:  ethkey.NewPrivKey(config.Wallet),
	})

	logger.Register(p.node, p.log)
	p.allowlist = allowlist.Register(p.node)
	p.banner = banner.Register(p.node)

	err = p.startNode()
	if err != nil {
		return nil, err
	}
	err = p.bannedAddrs(config.BannedAddrs)
	if err != nil {
		return nil, err
	}
	err = p.allowedPeers(config.AllowedPeers)
	if err != nil {
		return nil, err
	}
	err = p.bootstrapAddrs(config.BootstrapAddrs)
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
	sub, err := p.node.Subscription(topic)
	if err != nil {
		return err
	}
	return sub.Publish(message)
}

// WaitFor implements the transport.Transport interface.
func (p *P2P) WaitFor(topic string, message transport.Message) chan transport.Status {
	sub, err := p.node.Subscription(topic)
	if err != nil {
		return nil
	}
	return sub.Next(message)
}

// Close implements the transport.Transport interface.
func (p *P2P) Close() error {
	defer p.log.Info("Stopped")
	return p.node.Close()
}

// startNode starts libp2p node.
func (p *P2P) startNode() error {
	err := p.node.Start()
	if err != nil {
		return err
	}

	p.log.
		WithFields(log.Fields{"addrs": p.listenAddrs()}).
		Info("Listening")

	return nil
}

// bannedAddrs bans all addresses nodes from the addrs list using the
// banner package.
func (p *P2P) bannedAddrs(addrs []string) error {
	for _, addrstr := range addrs {
		maddr, err := multiaddr.NewMultiaddr(addrstr)
		if err != nil {
			return err
		}

		err = p.banner.Ban(maddr)
		if err != nil {
			return err
		}
	}
	return nil
}

// allowedPeers add peers to allowed list using the allowlist package. Only
// peers from that list will be allowed to send messages.
func (p *P2P) allowedPeers(ids []string) error {
	for _, idstr := range ids {
		id, err := peer.Decode(idstr)
		if err != nil {
			return err
		}
		p.allowlist.Allow(id)
	}
	return nil
}

// bootstrapAddrs connects to all nodes from the addrs list.
func (p *P2P) bootstrapAddrs(addrs []string) error {
	for _, addrstr := range addrs {
		maddr, err := multiaddr.NewMultiaddr(addrstr)
		if err != nil {
			return err
		}

		err = p.node.Connect(maddr)
		if err != nil {
			p.log.
				WithFields(log.Fields{"addr": addrstr}).
				WithError(err).
				Warn("Unable to connect to bootstrap peer")
		}
	}
	return nil
}

// listenAddrs returns all node's listen multiaddresses as a string list.
func (p *P2P) listenAddrs() []string {
	var strs []string
	for _, addr := range p.node.Host().Addrs() {
		strs = append(strs, fmt.Sprintf("%s/p2p/%s", addr.String(), p.node.Host().ID()))
	}
	return strs
}

// addrsToMaddrs converts multiaddresses given as strings to a
// list of multiaddr.Multiaddr.
func addrsToMaddrs(addrs []string) ([]multiaddr.Multiaddr, error) {
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
