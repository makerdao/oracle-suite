package p2p

import (
	"context"

	pubsub "github.com/libp2p/go-libp2p-pubsub"

	"github.com/makerdao/gofer/internal/log"
	"github.com/makerdao/gofer/internal/transport"
	"github.com/makerdao/gofer/internal/transport/p2p/sets"
)

type subscription struct {
	ctx            context.Context
	topic          *pubsub.Topic
	sub            *pubsub.Subscription
	teh            *pubsub.TopicEventHandler
	eventHandler   sets.EventHandler
	messageHandler sets.MessageHandler
	log            log.Logger

	// statusCh is used to send a notification about new message, it's returned by
	// transport.WaitFor function.
	statusCh chan transport.Status
}

func newSubscription(node *Node, topic string) (*subscription, error) {
	t, err := node.pubSub.Join(topic)
	if err != nil {
		return nil, err
	}
	s, err := t.Subscribe()
	if err != nil {
		return nil, err
	}
	teh, err := t.EventHandler()
	if err != nil {
		return nil, err
	}

	sub := &subscription{
		ctx:            node.ctx,
		topic:          t,
		sub:            s,
		teh:            teh,
		eventHandler:   node.eventHandlerSet,
		messageHandler: node.messageHandlerSet,
		statusCh:       make(chan transport.Status, 0),
	}

	sub.eventLoop()

	return sub, err
}

func (s subscription) publish(message transport.Message) error {
	b, err := message.Marshall()
	if err != nil {
		return err
	}
	s.messageHandler.Published(s.topic.String(), b, message)
	return s.topic.Publish(s.ctx, b)
}

func (s subscription) next(message transport.Message) chan transport.Status {
	go func() {
		msg, err := s.sub.Next(s.ctx)
		if err == nil {
			err = message.Unmarshall(msg.Data)
		}
		s.messageHandler.Received(s.topic.String(), msg, message)
		s.statusCh <- transport.Status{
			Message: message,
			Error:   err,
		}
	}()

	return s.statusCh
}

func (s subscription) eventLoop() {
	go func() {
		for {
			pe, err := s.teh.NextPeerEvent(s.ctx)
			if err != nil {
				// THe only situation when error may be returned here, is when
				// subscription is canceled.
				return
			}
			s.eventHandler.Handle(s.topic.String(), pe)
		}
	}()
}

func (s subscription) close() error {
	s.sub.Cancel()
	return s.topic.Close()
}
