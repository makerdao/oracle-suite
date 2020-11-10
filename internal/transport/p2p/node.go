package p2p

import (
	"context"
	"errors"
	"sync"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/connmgr"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/multiformats/go-multiaddr"

	"github.com/makerdao/gofer/internal/log"
	"github.com/makerdao/gofer/internal/transport/p2p/sets"
)

var ConnectionIsClosedErr = errors.New("connection is closed")
var AlreadySubscribedErr = errors.New("topic is already subscribed")
var TopicIsNotSubscribedErr = errors.New("topic is not subscribed")

type Node struct {
	mu sync.Mutex

	ctx               context.Context
	host              host.Host
	pubSub            *pubsub.PubSub
	notifeeSet        *sets.NotifeeSet
	connGaterSet      *sets.ConnGaterSet
	validatorSet      *sets.ValidatorSet
	eventHandlerSet   *sets.EventHandlerSet
	messageHandlerSet *sets.MessageHandlerSet
	subs              map[string]*subscription
	log               log.Logger
	closed            bool
}

func NewNode(ctx context.Context, l log.Logger) *Node {
	return &Node{
		ctx:               ctx,
		notifeeSet:        sets.NewNotifeeSet(),
		connGaterSet:      sets.NewConnGaterSet(),
		validatorSet:      sets.NewValidatorSet(),
		eventHandlerSet:   sets.NewEventHandlerSet(),
		messageHandlerSet: sets.NewMessageHandlerSet(),
		subs:              make(map[string]*subscription, 0),
		log:               l,
		closed:            false,
	}
}

func (n *Node) Start(maddrs []multiaddr.Multiaddr) error {
	h, err := libp2p.New(n.ctx,
		libp2p.ListenAddrs(maddrs...),
		libp2p.ConnectionGater(n.connGaterSet),
	)
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

func (n *Node) ID() peer.ID {
	return n.host.ID()
}

func (n *Node) Addrs() []multiaddr.Multiaddr {
	return n.host.Addrs()
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

func (n *Node) AddValidator(topic string, validator pubsub.ValidatorEx) {
	n.validatorSet.Add(topic, validator)
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
