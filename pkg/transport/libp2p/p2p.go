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

package libp2p

import (
	"context"
	"errors"
	"fmt"
	"time"

	core "github.com/libp2p/go-libp2p-core"
	"github.com/libp2p/go-libp2p-core/crypto"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/multiformats/go-multiaddr"

	"github.com/makerdao/oracle-suite/internal/libp2p"
	"github.com/makerdao/oracle-suite/pkg/ethereum"
	"github.com/makerdao/oracle-suite/pkg/log"
	"github.com/makerdao/oracle-suite/pkg/transport"
	"github.com/makerdao/oracle-suite/pkg/transport/messages"
)

const LoggerTag = "P2P"

// Values for a connection limiter:
const minConnections = 100
const maxConnections = 150

// Mode describes operating mode of a node.
type Mode int

const (
	// ClientMode operates the node as client. ClientMode can publish and read messages
	// and provides peer discovery service for other nodes.
	ClientMode Mode = iota
	// BootstrapMode operates the node as a bootstrap node. BootstrapMode node provide
	// only peer discovery service for other nodes.
	BootstrapMode
)

// Values for a peer scoring and rate limiter:
const maxBytesPerSecond float64 = 10 * 1024 * 1024 // 10MB/s
const maxInvalidMsgsPerHour float64 = 60
const maxMsgSizeInBytes float64 = 64 * 1024
const priceUpdateInterval = time.Minute

// defaultListenAddrs is a list of default multiaddresses on which node will
// be listening on.
var defaultListenAddrs = []string{"/ip4/0.0.0.0/tcp/0"}

// LibP2P is a little wrapper for the Node that implements the transport.Transport
// interface.
type LibP2P struct {
	node   *libp2p.Node
	topics map[string]transport.Message
}

// Config is a configuration for the LibP2P transport.
type Config struct {
	// Mode describes in what mode the node should operate.
	Mode Mode
	// Topics is a list of subscribed topics. A value of the map a type of
	// a message given as a nil pointer, e.g.: (*Message)(nil).
	Topics map[string]transport.Message
	// PeerPrivKey is a key used for peer identity. If empty, then random key
	// is used. Ignored in bootstrap mode.
	PeerPrivKey crypto.PrivKey
	// MessagePrivKey is a key used to sign messages. If empty, then message
	// are signed with the same key which is used for peer identity. Ignored
	// in bootstrap mode.
	MessagePrivKey crypto.PrivKey
	// ListenAddrs is a list of multiaddresses on which this node will be
	// listening on. If empty, the localhost, and a random port will be used.
	ListenAddrs []string
	// BootstrapAddrs is a list multiaddresses of initial peers to connect to.
	// This option is ignored when discovery is disabled.
	BootstrapAddrs []string
	// DirectPeersAddrs is a list multiaddresses of peers to which messages
	// will be send directly. This option has to be configured symmetrically
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
	// to connect to the network. Always enabled in bootstrap mode.
	Discovery bool
	// Signer used to verify price messages. Ignored in bootstrap mode.
	Signer ethereum.Signer
	// Logger is a custom logger instance. If not provided then null
	// logger is used.
	Logger log.Logger

	// Application info:
	AppName    string
	AppVersion string
}

// New returns a new instance of a transport, implemented with
// the libp2p library.
func New(ctx context.Context, cfg Config) (*LibP2P, error) {
	var err error

	if len(cfg.ListenAddrs) == 0 {
		cfg.ListenAddrs = defaultListenAddrs
	}
	if ctx == nil {
		return nil, errors.New("context must not be nil")
	}
	listenAddrs, err := strsToMaddrs(cfg.ListenAddrs)
	if err != nil {
		return nil, fmt.Errorf("P2P transport error, unable to parse listenAddrs: %w", err)
	}
	bootstrapAddrs, err := strsToMaddrs(cfg.BootstrapAddrs)
	if err != nil {
		return nil, fmt.Errorf("P2P transport error, unable to parse bootstrapAddrs: %w", err)
	}
	directPeersAddrs, err := strsToMaddrs(cfg.DirectPeersAddrs)
	if err != nil {
		return nil, fmt.Errorf("P2P transport error, unable to parse directPeersAddrs: %w", err)
	}
	blockedAddrs, err := strsToMaddrs(cfg.BlockedAddrs)
	if err != nil {
		return nil, fmt.Errorf("P2P transport error: unable to parse blockedAddrs: %w", err)
	}

	logger := cfg.Logger.WithField("tag", LoggerTag)
	opts := []libp2p.Options{
		libp2p.Logger(logger),
		libp2p.ConnectionLogger(),
		libp2p.PeerLogger(),
		libp2p.UserAgent(fmt.Sprintf("%s/%s", cfg.AppName, cfg.AppVersion)),
		libp2p.ListenAddrs(listenAddrs),
		libp2p.DirectPeers(directPeersAddrs),
		libp2p.Denylist(blockedAddrs),
		libp2p.ConnectionLimit(
			minConnections,
			maxConnections,
			5*time.Minute,
		),
		libp2p.Monitor(),
	}
	if cfg.PeerPrivKey != nil {
		opts = append(opts, libp2p.PeerPrivKey(cfg.PeerPrivKey))
	}
	switch cfg.Mode {
	case ClientMode:
		opts = append(opts,
			libp2p.MessageLogger(),
			libp2p.RateLimiter(rateLimiterConfig(cfg)),
			libp2p.PeerScoring(peerScoreParams, thresholds, func(topic string) *pubsub.TopicScoreParams {
				if topic == messages.PriceMessageName {
					return priceTopicScoreParams(cfg)
				}
				return nil
			}),
			oracle(cfg.FeedersAddrs, cfg.Signer, logger),
		)
		if cfg.MessagePrivKey != nil {
			opts = append(opts, libp2p.MessagePrivKey(cfg.MessagePrivKey))
		}
		if cfg.Discovery {
			opts = append(opts, libp2p.Discovery(bootstrapAddrs))
		}
	case BootstrapMode:
		cfg.Topics = nil
		opts = append(opts,
			libp2p.DisablePubSub(),
			libp2p.Discovery(bootstrapAddrs),
		)
	}

	n, err := libp2p.NewNode(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("P2P transport error, unable to initialize node: %w", err)
	}

	return &LibP2P{node: n, topics: cfg.Topics}, nil
}

// Start implements the transport.Transport interface.
func (p *LibP2P) Start() error {
	err := p.node.Start()
	if err != nil {
		return fmt.Errorf("P2P transport error, unable to start node: %w", err)
	}
	for topic, typ := range p.topics {
		err := p.subscribe(topic, typ)
		if err != nil {
			return err
		}
	}
	return nil
}

// Wait implements the transport.Transport interface.
func (p *LibP2P) Wait() {
	p.node.Wait()
}

// Broadcast implements the transport.Transport interface.
func (p *LibP2P) Broadcast(topic string, message transport.Message) error {
	sub, err := p.node.Subscription(topic)
	if err != nil {
		return fmt.Errorf("P2P transport error, unable to get subscription for %s topic: %w", topic, err)
	}
	return sub.Publish(message)
}

// Messages implements the transport.Transport interface.
func (p *LibP2P) Messages(topic string) chan transport.ReceivedMessage {
	sub, err := p.node.Subscription(topic)
	if err != nil {
		return nil
	}
	return sub.Next()
}

func (p *LibP2P) subscribe(topic string, typ transport.Message) error {
	err := p.node.Subscribe(topic, typ)
	if err != nil {
		return fmt.Errorf("P2P transport error, unable to subscribe to topic %s: %w", topic, err)
	}
	return nil
}

// strsToMaddrs converts multiaddresses given as strings to a
// list of multiaddr.Multiaddr.
func strsToMaddrs(addrs []string) ([]core.Multiaddr, error) {
	var maddrs []core.Multiaddr
	for _, addrstr := range addrs {
		maddr, err := multiaddr.NewMultiaddr(addrstr)
		if err != nil {
			return nil, err
		}
		maddrs = append(maddrs, maddr)
	}
	return maddrs, nil
}

func rateLimiterConfig(cfg Config) libp2p.RateLimiterConfig {
	bytesPerSecond := maxBytesPerSecond
	burstSize := maxBytesPerSecond * priceUpdateInterval.Seconds()
	return libp2p.RateLimiterConfig{
		BytesPerSecond:      maxBytesPerSecond / float64(len(cfg.FeedersAddrs)),
		BurstSize:           int(burstSize / float64(len(cfg.FeedersAddrs))),
		RelayBytesPerSecond: bytesPerSecond,
		RelayBurstSize:      int(burstSize),
	}
}
