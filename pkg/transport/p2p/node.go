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
	"sync"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/connmgr"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/transport"
	discovery "github.com/libp2p/go-libp2p-discovery"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	swarm "github.com/libp2p/go-libp2p-swarm"
	"github.com/multiformats/go-multiaddr"

	"github.com/makerdao/gofer/pkg/log"
	"github.com/makerdao/gofer/pkg/transport/p2p/allowlist"
	"github.com/makerdao/gofer/pkg/transport/p2p/denylist"
	"github.com/makerdao/gofer/pkg/transport/p2p/logger"
	"github.com/makerdao/gofer/pkg/transport/p2p/sets"
)

var ErrConnectionIsClosed = errors.New("connection is closed")
var ErrAlreadySubscribed = errors.New("topic is already subscribed")
var ErrTopicIsNotSubscribed = errors.New("topic is not subscribed")

const rendezvousString = "spire/0.0-dev"

func init() {
	// It's required to increase timeouts because signing messages using
	// the Ethereum wallet may take more time than default timeout allows.
	const timeout = 120 * time.Second
	transport.DialTimeout = timeout
	swarm.DialTimeoutLocal = timeout
}

type NodeConfig struct {
	Context context.Context
	Logger  log.Logger

	ListenAddrs    []multiaddr.Multiaddr
	BootstrapAddrs []multiaddr.Multiaddr
	BlockedAddrs   []multiaddr.Multiaddr
	AllowedPeers   []peer.ID
	PrivateKey     crypto.PrivKey
}

type Node struct {
	mu sync.Mutex

	ctx               context.Context
	host              host.Host
	pubSub            *pubsub.PubSub
	dht               *dht.IpfsDHT
	privKey           crypto.PrivKey
	listenAddrs       []multiaddr.Multiaddr
	bootstrapAddrs    []multiaddr.Multiaddr
	blockedAddrs      []multiaddr.Multiaddr
	allowedPeers      []peer.ID
	notifeeSet        *sets.NotifeeSet
	connGaterSet      *sets.ConnGaterSet
	validatorSet      *sets.ValidatorSet
	eventHandlerSet   *sets.EventHandlerSet
	messageHandlerSet *sets.MessageHandlerSet
	allowlist         *allowlist.Allowlist
	denylist          *denylist.Denylist
	subs              map[string]*subscription
	log               log.Logger
	closed            bool
}

func NewNode(cfg NodeConfig) *Node {
	return &Node{
		ctx:               cfg.Context,
		privKey:           cfg.PrivateKey,
		bootstrapAddrs:    cfg.BootstrapAddrs,
		listenAddrs:       cfg.ListenAddrs,
		blockedAddrs:      cfg.BlockedAddrs,
		allowedPeers:      cfg.AllowedPeers,
		notifeeSet:        sets.NewNotifeeSet(),
		connGaterSet:      sets.NewConnGaterSet(),
		validatorSet:      sets.NewValidatorSet(),
		eventHandlerSet:   sets.NewEventHandlerSet(),
		messageHandlerSet: sets.NewMessageHandlerSet(),
		subs:              make(map[string]*subscription),
		log:               cfg.Logger,
		closed:            false,
	}
}

func (n *Node) Start() error {
	n.log.Info("Starting")

	var err error

	// Options:
	opts := []libp2p.Option{
		libp2p.ListenAddrs(n.listenAddrs...),
		libp2p.ConnectionGater(n.connGaterSet),
	}
	if n.privKey != nil {
		opts = append(opts, libp2p.Identity(n.privKey))
	}

	// Systems:
	n.host, err = libp2p.New(n.ctx, opts...)
	if err != nil {
		return err
	}
	n.pubSub, err = pubsub.NewGossipSub(n.ctx, n.host)
	if err != nil {
		return err
	}
	n.dht, err = dht.New(n.ctx, n.host)
	if err != nil {
		return err
	}

	// Logger:
	logger.Register(n, n.log)

	// Allowlist:
	n.allowlist = allowlist.Register(n, n.log)
	for _, p := range n.allowedPeers {
		n.allowlist.Allow(p)
	}

	// Denylist:
	n.denylist = denylist.Register(n)
	for _, a := range n.blockedAddrs {
		err = n.denylist.Deny(a)
		if err != nil {
			n.log.
				WithError(err).
				WithField("maddr", a.String()).
				Error("Unable to add given address to denylist")
		}
	}

	n.host.Network().Notify(n.notifeeSet)

	// Bootstrap peers:
	for _, maddr := range n.bootstrapAddrs {
		err = n.Connect(maddr)
		if err != nil {
			n.log.
				WithFields(log.Fields{"addr": maddr.String()}).
				WithError(err).
				Warn("Unable to connect to bootstrap peer")
		}
	}

	// Use a rendezvous point to announce our location:
	routingDiscovery := discovery.NewRoutingDiscovery(n.dht)
	discovery.Advertise(n.ctx, routingDiscovery, rendezvousString)

	n.log.
		WithFields(log.Fields{"addrs": n.nodeListenAddrs()}).
		Info("Listening")

	return nil
}

func (n *Node) Stop() error {
	if n.closed {
		return ErrConnectionIsClosed
	}

	defer n.log.Info("Stopped")

	n.mu.Lock()
	defer n.mu.Unlock()
	var err error

	// Close subscriptions:
	for t, s := range n.subs {
		err = s.close()
		if err != nil {
			n.log.
				WithError(err).
				WithField("topic", t).
				Error("Unable to close subscription")
		}
	}

	// Close DHT:
	err = n.dht.Close()
	if err != nil {
		n.log.
			WithError(err).
			Error("Unable to close DHT")
	}

	n.subs = nil
	n.closed = true
	return n.host.Close()
}

func (n *Node) Host() host.Host {
	return n.host
}

func (n *Node) PubSub() *pubsub.PubSub {
	return n.pubSub
}

func (n *Node) Connect(maddr multiaddr.Multiaddr) error {
	pi, err := peer.AddrInfoFromP2pAddr(maddr)
	if err != nil {
		return err
	}
	err = n.host.Connect(n.ctx, *pi)
	if err != nil {
		return err
	}
	return nil
}

func (n *Node) AddNotifee(notifees ...network.Notifiee) {
	n.notifeeSet.Add(notifees...)
}

func (n *Node) AddConnectionGater(connGaters ...connmgr.ConnectionGater) {
	n.connGaterSet.Add(connGaters...)
}

func (n *Node) AddValidator(validator sets.Validator) {
	n.validatorSet.Add(validator)
}

func (n *Node) AddEventHandler(eventHandler ...sets.EventHandler) {
	n.eventHandlerSet.Add(eventHandler...)
}

func (n *Node) AddMessageHandler(messageHandlers ...sets.MessageHandler) {
	n.messageHandlerSet.Add(messageHandlers...)
}

func (n *Node) Subscribe(topic string) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.closed {
		return ErrConnectionIsClosed
	}
	if _, ok := n.subs[topic]; ok {
		return ErrAlreadySubscribed
	}

	err := n.pubSub.RegisterTopicValidator(topic, n.validatorSet.Validator(topic))
	if err != nil {
		return err
	}

	sub, err := newSubscription(n, topic)
	if err != nil {
		return err
	}
	n.subs[topic] = sub

	return nil
}

func (n *Node) Unsubscribe(topic string) error {
	if n.closed {
		return ErrConnectionIsClosed
	}

	sub, err := n.subscription(topic)
	if err != nil {
		return err
	}

	return sub.close()
}

func (n *Node) subscription(topic string) (*subscription, error) {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.closed {
		return nil, ErrConnectionIsClosed
	}
	if sub, ok := n.subs[topic]; ok {
		return sub, nil
	}

	return nil, ErrTopicIsNotSubscribed
}

// nodeListenAddrs returns all node's listen multiaddresses as a string list.
func (n *Node) nodeListenAddrs() []string {
	var strs []string
	for _, addr := range n.host.Addrs() {
		strs = append(strs, fmt.Sprintf("%s/p2p/%s", addr.String(), n.host.ID()))
	}
	return strs
}
