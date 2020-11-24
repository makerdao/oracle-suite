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

	pubsub "github.com/libp2p/go-libp2p-pubsub"

	"github.com/makerdao/gofer/pkg/transport"
	"github.com/makerdao/gofer/pkg/transport/p2p/sets"
)

type subscription struct {
	ctx            context.Context
	topic          *pubsub.Topic
	sub            *pubsub.Subscription
	teh            *pubsub.TopicEventHandler
	eventHandler   sets.EventHandler
	messageHandler sets.MessageHandler

	// statusCh is used to send a notification about a new message, it's
	// returned by the Transport.WaitFor function.
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
	e, err := t.EventHandler()
	if err != nil {
		return nil, err
	}

	sub := &subscription{
		ctx:            node.ctx,
		topic:          t,
		sub:            s,
		teh:            e,
		eventHandler:   node.eventHandlerSet,
		messageHandler: node.messageHandlerSet,
		statusCh:       make(chan transport.Status),
	}

	sub.eventLoop()
	return sub, err
}

func (s subscription) Publish(message transport.Message) error {
	b, err := message.Marshall()
	if err != nil {
		return err
	}
	s.messageHandler.Published(s.topic.String(), b, message)
	return s.topic.Publish(s.ctx, b)
}

func (s subscription) Next(message transport.Message) chan transport.Status {
	go func() {
		msg, err := s.sub.Next(s.ctx)
		if err == nil {
			err = message.Unmarshall(msg.Data)
		}
		if msg != nil {
			s.messageHandler.Received(s.topic.String(), msg, message)
		}
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
				// The only situation when an error may be returned here is
				// when the subscription is canceled.
				return
			}
			s.eventHandler.Handle(s.topic.String(), pe)
		}
	}()
}

func (s subscription) close() error {
	s.teh.Cancel()
	s.sub.Cancel()
	return s.topic.Close()
}
