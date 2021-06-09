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
	"errors"
	"fmt"
	"time"

	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/multiformats/go-multiaddr"

	"github.com/makerdao/oracle-suite/internal/p2p"
	"github.com/makerdao/oracle-suite/pkg/ethereum"
	"github.com/makerdao/oracle-suite/pkg/log"
	"github.com/makerdao/oracle-suite/pkg/transport"
)

const LoggerTag = "P2P"
const userAgentString = "spire/v0.0-dev"

// Values for connection limiter:
const lowPeers = 100
const highPeers = 150

// Values used to calculate limits for a rate limiter:
const maxMessageSize = 128 * 1024 // maximum expected message size in bytes
const maxPairs = 50               // maximum expected number of asset pairs
const priceUpdateInterval = 60    // expected price update interval in seconds

// defaultListenAddrs is a list of default multiaddresses on which node will
// be listening on.
var defaultListenAddrs = []string{"/ip4/0.0.0.0/tcp/0"}

var ErrP2P = errors.New("P2P transport error")

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
	// This option is ignored when discovery is disabled.
	BootstrapAddrs []string
	// DirectPeersAddrs is a list multiaddresses of peers to which messages
	// will be send directly. This option have to be configured symmetrically
	// at both ends.
	DirectPeersAddrs []string
	// BlockedAddrs is a list of multiaddresses to which connection will be
	// blocked. If an address on that list contains an IP and a peer ID, both
	// will be blocked separately.
	BlockedAddrs []string
	// FeedersAddrs is a list of price feeders. Only feeders can create new
	// messages in the network.
	FeedersAddrs []ethereum.Address
	// Discovery indicates whenever peer discovery should be enabled.
	// If discovery is disabled, then DirectPeersAddrs must be used
	// to connect to the network.
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
		return nil, fmt.Errorf("%v: unable to parse listenAddrs: %v", ErrP2P, err)
	}
	bootstrapAddrs, err := strsToMaddrs(cfg.BootstrapAddrs)
	if err != nil {
		return nil, fmt.Errorf("%v: unable to parse bootstrapAddrs: %v", ErrP2P, err)
	}
	directPeersAddrs, err := strsToMaddrs(cfg.DirectPeersAddrs)
	if err != nil {
		return nil, fmt.Errorf("%v: unable to parse directPeersAddrs: %v", ErrP2P, err)
	}
	blockedAddrs, err := strsToMaddrs(cfg.BlockedAddrs)
	if err != nil {
		return nil, fmt.Errorf("%v: unable to parse blockedAddrs: %v", ErrP2P, err)
	}

	logger := cfg.Logger.WithField("tag", LoggerTag)
	opts := []p2p.Options{
		p2p.Logger(logger),
		p2p.ConnectionLogger(),
		p2p.MessageLogger(),
		p2p.PeerLogger(),
		p2p.UserAgent(userAgentString),
		p2p.ListenAddrs(listenAddrs),
		p2p.DirectPeers(directPeersAddrs),
		p2p.Denylist(blockedAddrs),
		p2p.ConnectionLimit(lowPeers, highPeers, 5*time.Minute),
		p2p.RateLimiter(p2p.RateLimiterConfig{
			GlobalBytesPerSecond: maxMessageSize * maxPairs / priceUpdateInterval * float64(len(cfg.FeedersAddrs)),
			PeerBytesPerSecond:   maxMessageSize * maxPairs / priceUpdateInterval,
			GlobalBurst:          maxMessageSize * maxPairs * len(cfg.FeedersAddrs),
			PeerBurst:            maxMessageSize * maxPairs,
		}),
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
	}

	n, err := p2p.NewNode(cfg.Context, opts...)
	if err != nil {
		return nil, fmt.Errorf("%v: unable to initialize node: %v", ErrP2P, err)
	}

	p := &P2P{node: n}
	err = p.node.Start()
	if err != nil {
		return nil, fmt.Errorf("%v: unable to start node: %v", ErrP2P, err)
	}

	return p, nil
}

// Subscribe implements the transport.Transport interface.
func (p *P2P) Subscribe(topic string, typ transport.Message) error {
	err := p.node.Subscribe(topic, typ)
	if err != nil {
		return fmt.Errorf("%v: unable to subscribe to topic %s: %v", ErrP2P, topic, err)
	}
	return nil
}

// Unsubscribe implements the transport.Transport interface.
func (p *P2P) Unsubscribe(topic string) error {
	err := p.node.Unsubscribe(topic)
	if err != nil {
		return fmt.Errorf("%v: unable to unsubscribe from topic %s: %v", ErrP2P, topic, err)
	}
	return nil
}

// Broadcast implements the transport.Transport interface.
func (p *P2P) Broadcast(topic string, message transport.Message) error {
	sub, err := p.node.Subscription(topic)
	if err != nil {
		return fmt.Errorf("%v: unable to get subscription for %s topic: %v", ErrP2P, topic, err)
	}
	return sub.Publish(message)
}

// WaitFor implements the transport.Transport interface.
func (p *P2P) WaitFor(topic string) chan transport.ReceivedMessage {
	sub, err := p.node.Subscription(topic)
	if err != nil {
		return nil
	}
	go func() {

	}()
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
