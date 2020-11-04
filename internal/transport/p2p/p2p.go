package p2p

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/libp2p/go-libp2p-core/host"
	pubsub "github.com/libp2p/go-libp2p-pubsub"

	"github.com/makerdao/gofer/internal/logger"
	"github.com/makerdao/gofer/internal/transport"
)

const LoggerTag = "P2P"

type P2P struct {
	mu     sync.RWMutex
	ctx    context.Context
	logger logger.Logger

	// node is a current libp2p's node
	host host.Host
	// ps is a instance of PubSub system
	ps *pubsub.PubSub
	// subs is a list of subscription we are interested in
	subs map[string]subscription
	// closed determines whenever P2P is closed
	closed bool
}

type subscription struct {
	topic *pubsub.Topic
	sub   *pubsub.Subscription

	// statusCh is used to send a notification about new message, it's returned by
	// transport.WaitFor function.
	statusCh chan transport.Status
}

func (s subscription) unsubscribe() error {
	s.sub.Cancel()
	return s.topic.Close()
}

type Config struct {
	Context context.Context
	Listen  string
	Peers   []string
	Logger  logger.Logger
}

func NewP2P(cfg Config) (*P2P, error) {
	var err error

	listen := cfg.Listen
	if listen == "" {
		listen = "/ip4/0.0.0.0/tcp/0"
	}

	p := &P2P{
		ctx:    cfg.Context,
		logger: cfg.Logger,
		subs:   make(map[string]subscription, 0),
	}

	err = p.setupNode(cfg.Context, listen, cfg.Peers)
	if err != nil {
		return nil, err
	}

	err = p.setupGossip()
	if err != nil {
		return nil, err
	}

	err = p.setupDiscovery()
	if err != nil {
		return nil, err
	}

	p.logger.Info(LoggerTag, "Initialized, listening on addresses: %s", strings.Join(p.Addresses(), ", "))
	return p, nil
}

// Addresses returns a list of addresses on which we are listening.
func (p *P2P) Addresses() []string {
	p.mu.RLock()
	p.mu.RUnlock()

	var addresses []string
	for _, addr := range p.host.Addrs() {
		p.host.Addrs()
		addresses = append(addresses, fmt.Sprintf("%s/p2p/%s", addr.String(), p.host.ID().String()))
	}

	return addresses
}

// Subscribe implements the transport.Transport interface.
func (p *P2P) Subscribe(topic string) error {
	p.mu.Lock()
	p.mu.Unlock()

	if p.closed {
		return errors.New("p2p is already closed")
	}

	if _, ok := p.subs[topic]; ok {
		return fmt.Errorf("unable to subscirbe to the %s topic becasue is already subscribed", topic)
	}

	t, err := p.ps.Join(topic)
	if err != nil {
		return err
	}

	s, err := t.Subscribe()
	if err != nil {
		return err
	}

	p.subs[topic] = subscription{
		topic:    t,
		sub:      s,
		statusCh: make(chan transport.Status, 0),
	}

	p.logger.Info(LoggerTag, "Message \"%s\" subscribed", topic)
	return nil
}

// Unsubscribe implements the transport.Transport interface.
func (p *P2P) Unsubscribe(topic string) error {
	p.mu.Lock()
	p.mu.Unlock()

	if _, ok := p.subs[topic]; !ok {
		return fmt.Errorf("unable to unsubscirbe to the %s topic becasue is already unsubscribed", topic)
	}

	p.logger.Info(LoggerTag, "Message \"%s\" unsubscribed", topic)
	return p.subs[topic].unsubscribe()
}

// Broadcast implements the transport.Transport interface.
func (p *P2P) Broadcast(topic string, payload transport.Message) error {
	p.mu.RLock()
	p.mu.RUnlock()

	if p.closed {
		return errors.New("p2p is already closed")
	}

	if _, ok := p.subs[topic]; !ok {
		return fmt.Errorf("unable to broadcast to the %s topic because is not subscribed", topic)
	}

	bts, err := payload.Marshall()
	if err != nil {
		return err
	}

	p.logger.Debug(LoggerTag, "Message \"%s\" broadcasted: ", topic, bts)
	return p.subs[topic].topic.Publish(p.ctx, bts)
}

// WaitFor implements the transport.Transport interface.
func (p *P2P) WaitFor(topic string, payload transport.Message) chan transport.Status {
	p.mu.RLock()
	p.mu.RUnlock()

	if _, ok := p.subs[topic]; !ok {
		return nil
	}

	sub := p.subs[topic].sub
	go func() {
		msg, err := sub.Next(p.ctx)
		if err == nil {
			p.logger.Debug(LoggerTag, "Message \"%s\" received: %s", topic, msg.Data)
		}

		// Try to unmarshall payload ONLY if there is no error.
		if err == nil {
			err = payload.Unmarshall(msg.Data)
		}

		p.subs[topic].statusCh <- transport.Status{
			Payload: payload,
			Error:   err,
		}
	}()

	return p.subs[topic].statusCh
}

// Close implements the transport.Transport interface.
func (p *P2P) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, s := range p.subs {
		_ = s.unsubscribe()
	}

	p.subs = nil
	p.closed = true
	err := p.host.Close()
	p.logger.Info(LoggerTag, "Closed")

	return err
}
