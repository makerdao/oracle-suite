package p2p

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/libp2p/go-libp2p-core/host"
	pubsub "github.com/libp2p/go-libp2p-pubsub"

	"github.com/makerdao/gofer/pkg/transport"
)

type P2P struct {
	mu sync.RWMutex

	// node is a current libp2p's node
	node host.Host
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

func (s subscription) Unsubscribe() error {
	err := s.topic.Close()
	s.sub.Cancel()
	close(s.statusCh)
	return err
}

func NewP2P(listen string, peers []string) (*P2P, error) {
	var err error

	ctx := context.Background()

	if listen == "" {
		listen = "/ip4/0.0.0.0/tcp/0"
	}

	p := &P2P{
		subs: make(map[string]subscription, 0),
	}

	err = p.setupNode(ctx, listen, peers)
	if err != nil {
		return nil, err
	}

	err = p.setupGossip(ctx)
	if err != nil {
		return nil, err
	}

	err = p.setupDiscovery(ctx)
	if err != nil {
		return nil, err
	}

	return p, nil
}

func (p *P2P) Addresses() []string {
	p.mu.RLock()
	p.mu.RUnlock()

	var addresses []string
	for _, addr := range p.node.Addrs() {
		addresses = append(addresses, addr.String())
	}

	return addresses
}

// Broadcast implements the transport.Transport interface.
func (p *P2P) Broadcast(eventName string, payload transport.Event) error {
	p.mu.RLock()
	p.mu.RUnlock()

	if p.closed {
		return errors.New("p2p is already closed")
	}

	if _, ok := p.subs[eventName]; !ok {
		return fmt.Errorf("unable to broadcast to the %s topic because is not subscribed", eventName)
	}

	bts, err := payload.PayloadMarshall()
	if err != nil {
		return err
	}

	return p.subs[eventName].topic.Publish(context.Background(), bts)
}

// Subscribe implements the transport.Transport interface.
func (p *P2P) Subscribe(eventName string) error {
	p.mu.Lock()
	p.mu.Unlock()

	if p.closed {
		return errors.New("p2p is already closed")
	}

	if _, ok := p.subs[eventName]; ok {
		return fmt.Errorf("unable to subscirbe to the %s topic becasue is already subscribed", eventName)
	}

	topic, err := p.ps.Join(eventName)
	if err != nil {
		return err
	}

	sub, err := topic.Subscribe()
	if err != nil {
		return err
	}

	p.subs[eventName] = subscription{
		topic:    topic,
		sub:      sub,
		statusCh: make(chan transport.Status, 0),
	}

	return nil
}

// Unsubscribe implements the transport.Transport interface.
func (p *P2P) Unsubscribe(eventName string) error {
	p.mu.Lock()
	p.mu.Unlock()

	if _, ok := p.subs[eventName]; !ok {
		return fmt.Errorf("unable to unsubscirbe to the %s topic becasue is already unsubscribed", eventName)
	}

	return p.subs[eventName].Unsubscribe()
}

// WaitFor implements the transport.Transport interface.
func (p *P2P) WaitFor(eventName string, payload transport.Event) chan transport.Status {
	p.mu.RLock()
	p.mu.RUnlock()

	if _, ok := p.subs[eventName]; !ok {
		return nil
	}

	sub := p.subs[eventName].sub
	go func() {
		msg, err := sub.Next(context.Background())

		// Try to unmarshall payload ONLY if there is no error.
		if err == nil {
			err = payload.PayloadUnmarshall(msg.Data)
		}

		p.subs[eventName].statusCh <- transport.Status{
			Error: err,
		}
	}()

	return p.subs[eventName].statusCh
}

// Close implements the transport.Transport interface.
func (p *P2P) Close() error {
	p.mu.Lock()
	p.mu.Unlock()

	for _, s := range p.subs {
		_ = s.Unsubscribe()
	}

	p.subs = nil
	p.closed = true
	return p.node.Close()
}
