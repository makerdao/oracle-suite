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
	connmgr "github.com/libp2p/go-libp2p-connmgr"
	coreConnmgr "github.com/libp2p/go-libp2p-core/connmgr"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/peerstore"
	"github.com/libp2p/go-libp2p-core/transport"
	"github.com/libp2p/go-libp2p-peerstore/pstoremem"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	swarm "github.com/libp2p/go-libp2p-swarm"
	"github.com/multiformats/go-multiaddr"

	"github.com/makerdao/oracle-suite/internal/p2p/sets"

	"github.com/makerdao/oracle-suite/pkg/log"
	"github.com/makerdao/oracle-suite/pkg/log/null"
	pkgTransport "github.com/makerdao/oracle-suite/pkg/transport"
)

var ErrConnectionClosed = errors.New("connection is closed")
var ErrAlreadySubscribed = errors.New("topic is already subscribed")
var ErrNotSubscribed = errors.New("topic is not subscribed")
var ErrPubSubDisabled = errors.New("pubsub protocol is disabled")

func init() {
	// It's required to increase timeouts because signing messages using
	// the Ethereum wallet may take more time than default timeout allows.
	const timeout = 120 * time.Second
	transport.DialTimeout = timeout
	swarm.DialTimeoutLocal = timeout
}

// Node is a single node in the P2P network. It wraps the libp2p library to
// provide an easier to use and use-case agnostic interface for the pubsub
// system.
type Node struct {
	ctx    context.Context
	mu     sync.Mutex
	doneCh chan struct{} // doneCh is used to unblock the Wait method.

	host                  host.Host
	pubSub                *pubsub.PubSub
	peerstore             peerstore.Peerstore
	connmgr               coreConnmgr.ConnManager
	nodeEventHandler      *sets.NodeEventHandlerSet
	pubSubEventHandlerSet *sets.PubSubEventHandlerSet
	notifeeSet            *sets.NotifeeSet
	connGaterSet          *sets.ConnGaterSet
	validatorSet          *sets.ValidatorSet
	messageHandlerSet     *sets.MessageHandlerSet
	subs                  map[string]*Subscription
	log                   log.Logger
	disablePubSub         bool
	closed                bool

	hostOpts   []libp2p.Option
	pubsubOpts []pubsub.Option
}

func NewNode(ctx context.Context, opts ...Options) (*Node, error) {
	n := &Node{
		ctx:                   ctx,
		doneCh:                make(chan struct{}),
		peerstore:             pstoremem.NewPeerstore(),
		nodeEventHandler:      sets.NewNodeEventHandlerSet(),
		pubSubEventHandlerSet: sets.NewPubSubEventHandlerSet(),
		notifeeSet:            sets.NewNotifeeSet(),
		connGaterSet:          sets.NewConnGaterSet(),
		validatorSet:          sets.NewValidatorSet(),
		messageHandlerSet:     sets.NewMessageHandlerSet(),
		subs:                  make(map[string]*Subscription),
		log:                   null.New(),
		closed:                false,
	}

	// Apply options:
	for _, opt := range opts {
		err := opt(n)
		if err != nil {
			return nil, fmt.Errorf("libp2p node error, unable to apply option: %w", err)
		}
	}

	if n.connmgr == nil {
		n.connmgr = connmgr.NewConnManager(0, 0, 5*time.Minute)
	}

	n.nodeEventHandler.Handle(sets.NodeConfiguredEvent{})

	return n, nil
}

func (n *Node) Start() error {
	n.log.Info("Starting")
	var err error

	n.nodeEventHandler.Handle(sets.NodeStartingEvent{})

	go n.contextCancelHandler()

	n.host, err = libp2p.New(n.ctx, append([]libp2p.Option{
		libp2p.EnableNATService(),
		libp2p.DisableRelay(),
		libp2p.Peerstore(n.peerstore),
		libp2p.ConnectionGater(n.connGaterSet),
		libp2p.ConnectionManager(n.connmgr),
	}, n.hostOpts...)...)
	if err != nil {
		return fmt.Errorf("libp2p node error, unable to initialize libp2p: %v", err)
	}
	n.host.Network().Notify(n.notifeeSet)

	n.log = n.log.WithField("x-hostID", n.host.ID().String())
	n.nodeEventHandler.Handle(sets.NodeHostStartedEvent{})

	n.log.
		WithField("listenAddrs", n.listenAddrStrs()).
		Info("Listening")

	if !n.disablePubSub {
		n.pubSub, err = pubsub.NewGossipSub(n.ctx, n.host, n.pubsubOpts...)
		if err != nil {
			return fmt.Errorf("libp2p node error, unable to initialize gosspib pubsub: %v", err)
		}
		n.nodeEventHandler.Handle(sets.NodePubSubStartedEvent{})
	}

	n.nodeEventHandler.Handle(sets.NodeStartedEvent{})

	return nil
}

func (n *Node) Wait() {
	<-n.doneCh
}

func (n *Node) Addrs() []multiaddr.Multiaddr {
	var addrs []multiaddr.Multiaddr
	for _, s := range n.listenAddrStrs() {
		addrs = append(addrs, multiaddr.StringCast(s))
	}
	return addrs
}

func (n *Node) Host() host.Host {
	return n.host
}

func (n *Node) PubSub() *pubsub.PubSub {
	return n.pubSub
}

func (n *Node) Peerstore() peerstore.Peerstore {
	return n.peerstore
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

func (n *Node) AddNodeEventHandler(eventHandler ...sets.NodeEventHandler) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.nodeEventHandler.Add(eventHandler...)
}

func (n *Node) AddPubSubEventHandler(eventHandler ...sets.PubSubEventHandler) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.pubSubEventHandlerSet.Add(eventHandler...)
}

func (n *Node) AddNotifee(notifees ...network.Notifiee) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.notifeeSet.Add(notifees...)
}

func (n *Node) AddConnectionGater(connGaters ...coreConnmgr.ConnectionGater) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.connGaterSet.Add(connGaters...)
}

func (n *Node) AddValidator(validator sets.Validator) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.validatorSet.Add(validator)
}

func (n *Node) AddMessageHandler(messageHandlers ...sets.MessageHandler) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.messageHandlerSet.Add(messageHandlers...)
}

func (n *Node) Subscribe(topic string, typ pkgTransport.Message) error {
	if n.pubSub == nil {
		return ErrPubSubDisabled
	}
	defer n.nodeEventHandler.Handle(sets.NodeTopicSubscribedEvent{Topic: topic})

	n.mu.Lock()
	defer n.mu.Unlock()

	if n.closed {
		return fmt.Errorf("libp2p node error: %v", ErrConnectionClosed)
	}
	if _, ok := n.subs[topic]; ok {
		return fmt.Errorf("libp2p node error: %v", ErrAlreadySubscribed)
	}

	sub, err := newSubscription(n, topic, typ)
	if err != nil {
		return err
	}
	n.subs[topic] = sub

	return nil
}

func (n *Node) Unsubscribe(topic string) error {
	if n.pubSub == nil {
		return ErrPubSubDisabled
	}
	if n.closed {
		return fmt.Errorf("libp2p node error: %w", ErrConnectionClosed)
	}

	defer n.nodeEventHandler.Handle(sets.NodeTopicUnsubscribedEvent{Topic: topic})

	sub, err := n.Subscription(topic)
	if err != nil {
		return err
	}

	return sub.close()
}

func (n *Node) Subscription(topic string) (*Subscription, error) {
	if n.pubSub == nil {
		return nil, ErrPubSubDisabled
	}
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.closed {
		return nil, fmt.Errorf("libp2p node error: %w", ErrConnectionClosed)
	}
	if sub, ok := n.subs[topic]; ok {
		return sub, nil
	}

	return nil, fmt.Errorf("libp2p node error: %w", ErrNotSubscribed)
}

// contextCancelHandler handles context cancellation.
func (n *Node) contextCancelHandler() {
	defer func() { close(n.doneCh) }()
	defer n.log.Info("Stopped")
	defer n.nodeEventHandler.Handle(sets.NodeStoppedEvent{})
	<-n.ctx.Done()

	n.nodeEventHandler.Handle(sets.NodeStoppingEvent{})

	n.mu.Lock()
	defer n.mu.Unlock()

	n.subs = nil
	n.closed = true
	err := n.host.Close()
	if err != nil {
		n.log.WithError(err).Error("Error during closing host")
	}
}

// ListenAddrs returns all node's listen multiaddresses as a string list.
func (n *Node) listenAddrStrs() []string {
	var strs []string
	for _, addr := range n.host.Addrs() {
		strs = append(strs, fmt.Sprintf("%s/p2p/%s", addr.String(), n.host.ID()))
	}
	return strs
}
