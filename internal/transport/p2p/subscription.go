package p2p

import (
	"context"

	pubsub "github.com/libp2p/go-libp2p-pubsub"

	"github.com/makerdao/gofer/internal/log"
	"github.com/makerdao/gofer/internal/transport"
)

type subscription struct {
	ctx   context.Context
	topic *pubsub.Topic
	sub   *pubsub.Subscription
	log   log.Logger

	// statusCh is used to send a notification about new message, it's returned by
	// transport.WaitFor function.
	statusCh chan transport.Status
}

func newSubscription(node *node, topic string) (*subscription, error) {
	t, err := node.pubSub.Join(topic)
	if err != nil {
		return nil, err
	}

	s, err := t.Subscribe()
	if err != nil {
		return nil, err
	}

	return &subscription{
		ctx:      node.ctx,
		topic:    t,
		sub:      s,
		statusCh: make(chan transport.Status, 0),
	}, err
}

func (s subscription) publish(message transport.Message) error {
	b, err := message.Marshall()
	if err != nil {
		return err
	}

	return s.topic.Publish(s.ctx, b)
}

func (s subscription) next(message transport.Message) chan transport.Status {
	go func() {
		msg, err := s.sub.Next(s.ctx)
		if err == nil {
			err = message.Unmarshall(msg.Data)
		}

		s.statusCh <- transport.Status{
			Message: message,
			Error:   err,
		}
	}()

	return s.statusCh
}

func (s subscription) close() error {
	s.sub.Cancel()
	return s.topic.Close()
}
