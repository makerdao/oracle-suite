package p2p

import (
	"context"
	"errors"
	"net"
	"sync"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/multiformats/go-multiaddr"

	"github.com/makerdao/gofer/internal/log"
)

var ConnectionIsClosedErr = errors.New("connection is closed")
var AlreadySubscribedErr = errors.New("topic is already subscribed")
var TopicIsNotSubscribedErr = errors.New("topic is not subscribed")

type node struct {
	mu sync.Mutex

	ctx       context.Context
	host      host.Host
	pubSub    *pubsub.PubSub
	notifee   *notifee
	connGater *connGater
	subs      map[string]*subscription
	log       log.Logger
	closed    bool
}

func newNode(ctx context.Context, l log.Logger) *node {
	return &node{
		ctx:       ctx,
		notifee:   newNotifee(l),
		connGater: newConnGater(l),
		subs:      make(map[string]*subscription, 0),
		log:       l,
		closed:    false,
	}
}

func (n *node) start(maddrs []multiaddr.Multiaddr) error {
	h, err := libp2p.New(n.ctx,
		libp2p.ListenAddrs(maddrs...),
		libp2p.ConnectionGater(n.connGater),
	)
	if err != nil {
		return err
	}

	ps, err := pubsub.NewGossipSub(n.ctx, h)
	if err != nil {
		return err
	}

	h.Network().Notify(n.notifee)

	n.host = h
	n.pubSub = ps
	return nil
}

func (n *node) id() peer.ID {
	return n.host.ID()
}

func (n *node) addrs() []multiaddr.Multiaddr {
	return n.host.Addrs()
}

func (n *node) connect(maddr multiaddr.Multiaddr) error {
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

func (n *node) ban(maddr multiaddr.Multiaddr) error {
	multiaddr.ForEach(maddr, func(c multiaddr.Component) bool {
		switch c.Protocol().Code {
		case multiaddr.P_IP4:
			n.connGater.bannedAddrs.AddFilter(net.IPNet{
				IP:   net.ParseIP(c.String()),
				Mask: net.CIDRMask(32, 32),
			}, multiaddr.ActionDeny)
		case multiaddr.P_IP6:
			n.connGater.banIP(net.ParseIP(c.String()))
		case multiaddr.P_P2P:
			id, err := peer.IDFromBytes(c.RawValue())
			if err != nil {
				return true
			}
			n.pubSub.BlacklistPeer(id)
		}
		return true
	})

	return nil
}

func (n *node) subscription(topic string) (*subscription, error) {
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

func (n *node) subscribe(topic string) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.closed {
		return ConnectionIsClosedErr
	}
	if _, ok := n.subs[topic]; ok {
		return AlreadySubscribedErr
	}

	sub, err := newSubscription(n, topic)
	if err != nil {
		return err
	}
	n.subs[topic] = sub

	return nil
}

func (n *node) unsubscribe(topic string) error {
	n.mu.Lock()
	n.mu.Unlock()

	if n.closed {
		return ConnectionIsClosedErr
	}

	sub, err := n.subscription(topic)
	if err != nil {
		return err
	}

	return sub.close()
}

func (n *node) close() error {
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
