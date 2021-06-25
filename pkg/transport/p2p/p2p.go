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
	"math"
	"time"

	core "github.com/libp2p/go-libp2p-core"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/multiformats/go-multiaddr"

	"github.com/makerdao/oracle-suite/internal/p2p"
	"github.com/makerdao/oracle-suite/pkg/ethereum"
	"github.com/makerdao/oracle-suite/pkg/log"
	"github.com/makerdao/oracle-suite/pkg/transport"
	"github.com/makerdao/oracle-suite/pkg/transport/messages"
)

const LoggerTag = "P2P"

// Values for a connection limiter:
var lowPeers = 100
var highPeers = 150

// Values for a peer scoring and rate limiter:
const maxBytesPerSecond float64 = 10 * 1024 * 1024 // 10MB/s
const maxInvalidMessageCountPerHour float64 = 60
const maxMessageSizeInBytes float64 = 64 * 1024
const priceUpdateIntervalInSeconds float64 = 60

// defaultListenAddrs is a list of default multiaddresses on which node will
// be listening on.
var defaultListenAddrs = []string{"/ip4/0.0.0.0/tcp/0"}

var ErrP2P = errors.New("P2P transport error")

// P2P is a little wrapper for the Node that implements the transport.Transport
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

	// Application info:
	AppName    string
	AppVersion string
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
		p2p.UserAgent(fmt.Sprintf("%s/%s", cfg.AppName, cfg.AppVersion)),
		p2p.ListenAddrs(listenAddrs),
		p2p.DirectPeers(directPeersAddrs),
		p2p.Denylist(blockedAddrs),
		p2p.ConnectionLimit(
			lowPeers,
			highPeers,
			5*time.Minute,
		),
		p2p.RateLimiter(rateLimiter(cfg)),
		p2p.PeerScoring(peerScoreParams, thresholds, func(topic string) *pubsub.TopicScoreParams {
			if topic == messages.PriceMessageName {
				return priceTopicScoreParams(cfg)
			}
			return nil
		}),
		oracle(cfg.FeedersAddrs, cfg.Signer, logger),
		p2p.Monitor(),
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
	return sub.Next()
}

// Close implements the transport.Transport interface.
func (p *P2P) Close() error {
	return p.node.Stop()
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

// Peer scoring:
// https://github.com/libp2p/specs/blob/master/pubsub/gossipsub/gossipsub-v1.1.md#peer-scoring
const decayIntervalInSeconds float64 = 60
const decayToZero float64 = 0.01
const maxP1Points float64 = 500
const maxP2Points float64 = 1000
const maxP3Points float64 = -1000
const p1LengthInSeconds float64 = 3600
const p2LengthInSeconds float64 = 3600
const p3LengthInSeconds float64 = 5 * 60
const maxPoints = maxP1Points + maxP2Points

//nolint:gomnd
var thresholds = &pubsub.PeerScoreThresholds{
	GossipThreshold:   -250,
	PublishThreshold:  -500,
	GraylistThreshold: -750,
	AcceptPXThreshold: 0,
}

//nolint:gomnd
var peerScoreParams = &pubsub.PeerScoreParams{
	AppSpecificScore:            func(id peer.ID) float64 { return 0 },
	AppSpecificWeight:           1,
	IPColocationFactorWeight:    -1,
	IPColocationFactorThreshold: 1,
	BehaviourPenaltyWeight:      -10,
	BehaviourPenaltyThreshold:   1,
	BehaviourPenaltyDecay:       decay(maxPoints, 3600),
	DecayInterval:               time.Duration(decayIntervalInSeconds) * time.Second,
	DecayToZero:                 decayToZero,
	RetainScore:                 5 * time.Minute,
	Topics:                      make(map[string]*pubsub.TopicScoreParams),
}

func priceTopicScoreParams(cfg Config) *pubsub.TopicScoreParams {
	var maxPeers = float64(pubsub.GossipSubDhi)
	var minFeederCount = float64(len(cfg.FeedersAddrs)) / 2 // assume that 50% of feeders are offline
	var maxFeederCount = float64(len(cfg.FeedersAddrs))
	var minAssetPairCount = float64(1)
	var maxAssetPairCount = maxBytesPerSecond / maxMessageSizeInBytes * priceUpdateIntervalInSeconds / maxFeederCount
	var minMsgsPerSecond = (minFeederCount * minAssetPairCount) / maxPeers / priceUpdateIntervalInSeconds
	var maxMsgsPerSecond = (maxFeederCount * maxAssetPairCount) / priceUpdateIntervalInSeconds

	//nolint:gomnd
	return &pubsub.TopicScoreParams{
		TopicWeight: 1,

		// P₁
		TimeInMeshWeight:  maxP1Points / p1LengthInSeconds,
		TimeInMeshQuantum: time.Second,
		TimeInMeshCap:     p1LengthInSeconds,

		// P₂
		FirstMessageDeliveriesWeight: maxP2Points / (minMsgsPerSecond * p2LengthInSeconds),
		FirstMessageDeliveriesDecay:  decay(minMsgsPerSecond*p2LengthInSeconds, p2LengthInSeconds),
		FirstMessageDeliveriesCap:    minMsgsPerSecond * p2LengthInSeconds,

		// P₃
		// Note that because of the difference between the cap value and the threshold value, the decay time may be
		// longer than p3LengthInSeconds. The p3LengthInSeconds parameter describes the decay time for the minimum
		// number of messages.
		MeshMessageDeliveriesWeight:     maxP3Points / math.Pow(minMsgsPerSecond*p3LengthInSeconds, 2),
		MeshMessageDeliveriesDecay:      decay(minMsgsPerSecond*p3LengthInSeconds, p3LengthInSeconds),
		MeshMessageDeliveriesCap:        maxMsgsPerSecond * p3LengthInSeconds,
		MeshMessageDeliveriesThreshold:  minMsgsPerSecond * p3LengthInSeconds,
		MeshMessageDeliveriesActivation: time.Duration(p3LengthInSeconds) * time.Second,
		MeshMessageDeliveriesWindow:     10 * time.Millisecond,
		MeshFailurePenaltyWeight:        maxP3Points / math.Pow(minMsgsPerSecond*p3LengthInSeconds, 2),
		MeshFailurePenaltyDecay:         decay(minMsgsPerSecond*p3LengthInSeconds, p3LengthInSeconds),

		// P₄
		InvalidMessageDeliveriesWeight: maxPoints / maxInvalidMessageCountPerHour * -1,
		InvalidMessageDeliveriesDecay:  decay(maxPoints, 3600),
	}
}

// decay calculates a decay parameter for a peer scoring. It finds a number X
// that satisfies the equation: from*X^intervals=target. In other words, it
// finds a decay value for which a scoring will drop to the target value after
// the given number of seconds.
func decay(from float64, seconds float64) float64 {
	return math.Pow(decayToZero/from, 1/(seconds/decayIntervalInSeconds))
}

// Rate limiter:

func rateLimiter(cfg Config) p2p.RateLimiterConfig {
	bytesPerSecond := maxBytesPerSecond
	burstSize := maxBytesPerSecond * priceUpdateIntervalInSeconds
	return p2p.RateLimiterConfig{
		BytesPerSecond:      maxBytesPerSecond / float64(len(cfg.FeedersAddrs)),
		BurstSize:           int(burstSize / float64(len(cfg.FeedersAddrs))),
		RelayBytesPerSecond: bytesPerSecond,
		RelayBurstSize:      int(burstSize),
	}
}
