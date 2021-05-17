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

package node

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
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	swarm "github.com/libp2p/go-libp2p-swarm"
	"github.com/multiformats/go-multiaddr"

	"github.com/makerdao/oracle-suite/pkg/log"
	"github.com/makerdao/oracle-suite/pkg/log/null"
	pkgTransport "github.com/makerdao/oracle-suite/pkg/transport"
	"github.com/makerdao/oracle-suite/pkg/transport/p2p/node/sets"
)

var ErrConnectionClosed = errors.New("connection is closed")
var ErrAlreadySubscribed = errors.New("topic is already subscribed")
var ErrNotSubscribed = errors.New("topic is not subscribed")

func init() {
	// It's required to increase timeouts because signing messages using
	// the Ethereum wallet may take more time than default timeout allows.
	const timeout = 120 * time.Second
	transport.DialTimeout = timeout
	swarm.DialTimeoutLocal = timeout
}

type Node struct {
	mu sync.Mutex

	ctx                   context.Context
	host                  host.Host
	pubSub                *pubsub.PubSub
	peerPrivKey           crypto.PrivKey
	messagePrivKey        crypto.PrivKey
	messageAuthorPID      peer.ID
	listenAddrs           []multiaddr.Multiaddr
	nodeEventHandler      *sets.NodeEventHandlerSet
	pubSubEventHandlerSet *sets.PubSubEventHandlerSet
	notifeeSet            *sets.NotifeeSet
	connGaterSet          *sets.ConnGaterSet
	validatorSet          *sets.ValidatorSet
	messageHandlerSet     *sets.MessageHandlerSet
	subs                  map[string]*Subscription
	log                   log.Logger
	closed                bool

	newHost   func(n *Node) (host.Host, error)
	newPubSub func(n *Node) (*pubsub.PubSub, error)
}

func NewNode(ctx context.Context, opts ...Options) (*Node, error) {
	n := &Node{
		ctx:                   ctx,
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
			return nil, err
		}
	}

	// Systems providers:
	n.newHost = func(n *Node) (host.Host, error) {
		opts := []libp2p.Option{
			libp2p.ListenAddrs(n.listenAddrs...),
			libp2p.ConnectionGater(n.connGaterSet),
		}
		// If peerPrivKey is set, use it as peer identity:
		if n.peerPrivKey != nil {
			opts = append(opts, libp2p.Identity(n.peerPrivKey))
		}
		h, err := libp2p.New(n.ctx, opts...)
		if err != nil {
			return nil, err
		}
		// If the messagePrivKey is set, we have to add this key to the peerstore,
		// otherwise it'll be impossible to use it to sign messages:
		if n.messagePrivKey != nil {
			err = h.Peerstore().AddPrivKey(n.messageAuthorPID, n.messagePrivKey)
			if err != nil {
				return nil, err
			}
		}
		return h, nil
	}
	n.newPubSub = func(n *Node) (*pubsub.PubSub, error) {
		var opts []pubsub.Option
		if n.messagePrivKey != nil {
			opts = append(opts, pubsub.WithMessageAuthor(n.messageAuthorPID))
		}
		return pubsub.NewGossipSub(n.ctx, n.host, opts...)
	}

	n.nodeEventHandler.Handle(sets.NodeConfigured)

	return n, nil
}

func (n *Node) Start() error {
	n.log.Info("Starting")
	var err error

	n.nodeEventHandler.Handle(sets.NodeStarting)
	defer n.nodeEventHandler.Handle(sets.NodeStarted)

	// Systems:
	n.host, err = n.newHost(n)
	if err != nil {
		return err
	}
	n.pubSub, err = n.newPubSub(n)
	if err != nil {
		return err
	}

	n.host.Network().Notify(n.notifeeSet)

	n.log.
		WithField("addrs", n.listenAddrStrs()).
		Info("Listening")

	return nil
}

func (n *Node) Stop() error {
	if n.closed {
		return ErrConnectionClosed
	}

	n.nodeEventHandler.Handle(sets.NodeStopping)
	defer n.log.Info("Stopped")
	defer n.nodeEventHandler.Handle(sets.NodeStopped)

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

func (n *Node) AddNodeEventHandler(eventHandler ...sets.NodeEventHandler) {
	n.nodeEventHandler.Add(eventHandler...)
}

func (n *Node) AddPubSubEventHandler(eventHandler ...sets.PubSubEventHandler) {
	n.pubSubEventHandlerSet.Add(eventHandler...)
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

func (n *Node) AddMessageHandler(messageHandlers ...sets.MessageHandler) {
	n.messageHandlerSet.Add(messageHandlers...)
}

func (n *Node) Subscribe(topic string, typ pkgTransport.Message) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.closed {
		return ErrConnectionClosed
	}
	if _, ok := n.subs[topic]; ok {
		return ErrAlreadySubscribed
	}

	sub, err := newSubscription(n, topic, typ)
	if err != nil {
		return err
	}
	n.subs[topic] = sub

	return nil
}

func (n *Node) Unsubscribe(topic string) error {
	if n.closed {
		return ErrConnectionClosed
	}

	sub, err := n.Subscription(topic)
	if err != nil {
		return err
	}

	return sub.close()
}

func (n *Node) Subscription(topic string) (*Subscription, error) {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.closed {
		return nil, ErrConnectionClosed
	}
	if sub, ok := n.subs[topic]; ok {
		return sub, nil
	}

	return nil, ErrNotSubscribed
}

// ListenAddrs returns all node's listen multiaddresses as a string list.
func (n *Node) listenAddrStrs() []string {
	var strs []string
	for _, addr := range n.host.Addrs() {
		strs = append(strs, fmt.Sprintf("%s/p2p/%s", addr.String(), n.host.ID()))
	}
	return strs
}
