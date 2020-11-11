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
	"sync"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/connmgr"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/transport"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	swarm "github.com/libp2p/go-libp2p-swarm"
	"github.com/multiformats/go-multiaddr"

	"github.com/makerdao/gofer/internal/log"
	"github.com/makerdao/gofer/internal/transport/p2p/sets"
)

var ConnectionIsClosedErr = errors.New("connection is closed")
var AlreadySubscribedErr = errors.New("topic is already subscribed")
var TopicIsNotSubscribedErr = errors.New("topic is not subscribed")

func init() {
	// It's required to increase timeouts because signing messages using
	// the Ethereum wallet may take more time than default timeout allows:
	transport.DialTimeout = 120 * time.Second
	swarm.DialTimeoutLocal = 120 * time.Second
}

type NodeConfig struct {
	Context context.Context
	Logger  log.Logger

	// ListenAddrs is a list of multi addresses on which node will be
	// listening on. If empty, localhost and random port will be used.
	ListenAddrs []multiaddr.Multiaddr
	// PrivateKey is a key used to identify itself and for signing published
	// messages.
	PrivateKey crypto.PrivKey
}

type Node struct {
	mu sync.Mutex

	ctx               context.Context
	host              host.Host
	pubSub            *pubsub.PubSub
	privKey           crypto.PrivKey
	listenAddrs       []multiaddr.Multiaddr
	notifeeSet        *sets.NotifeeSet
	connGaterSet      *sets.ConnGaterSet
	validatorSet      *sets.ValidatorSet
	eventHandlerSet   *sets.EventHandlerSet
	messageHandlerSet *sets.MessageHandlerSet
	subs              map[string]*subscription
	log               log.Logger
	closed            bool
}

func NewNode(config NodeConfig) *Node {
	return &Node{
		ctx:               config.Context,
		privKey:           config.PrivateKey,
		listenAddrs:       config.ListenAddrs,
		notifeeSet:        sets.NewNotifeeSet(),
		connGaterSet:      sets.NewConnGaterSet(),
		validatorSet:      sets.NewValidatorSet(),
		eventHandlerSet:   sets.NewEventHandlerSet(),
		messageHandlerSet: sets.NewMessageHandlerSet(),
		subs:              make(map[string]*subscription, 0),
		log:               config.Logger,
		closed:            false,
	}
}

func (n *Node) Start() error {
	opts := []libp2p.Option{
		libp2p.ListenAddrs(n.listenAddrs...),
		libp2p.ConnectionGater(n.connGaterSet),
	}
	if n.privKey != nil {
		opts = append(opts, libp2p.Identity(n.privKey))
	}

	h, err := libp2p.New(n.ctx, opts...)
	if err != nil {
		return err
	}

	ps, err := pubsub.NewGossipSub(n.ctx, h)
	if err != nil {
		return err
	}

	h.Network().Notify(n.notifeeSet)
	n.host = h
	n.pubSub = ps
	return nil
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

func (n *Node) Subscription(topic string) (*subscription, error) {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.closed {
		return nil, ConnectionIsClosedErr
	}
	if sub, ok := n.subs[topic]; ok {
		return sub, nil
	}

	return nil, TopicIsNotSubscribedErr
}

func (n *Node) Subscribe(topic string) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.closed {
		return ConnectionIsClosedErr
	}
	if _, ok := n.subs[topic]; ok {
		return AlreadySubscribedErr
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
	n.mu.Lock()
	n.mu.Unlock()

	if n.closed {
		return ConnectionIsClosedErr
	}

	sub, err := n.Subscription(topic)
	if err != nil {
		return err
	}

	return sub.close()
}

func (n *Node) Close() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.closed {
		return ConnectionIsClosedErr
	}

	for _, s := range n.subs {
		_ = s.close()
	}

	n.subs = nil
	n.closed = true

	return n.host.Close()
}
