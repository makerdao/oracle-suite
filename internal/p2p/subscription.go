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
	"reflect"

	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"

	"github.com/makerdao/oracle-suite/internal/p2p/sets"
	"github.com/makerdao/oracle-suite/pkg/transport"
)

var ErrNilMessage = errors.New("message is nil")

type Subscription struct {
	ctx            context.Context
	topic          *pubsub.Topic
	sub            *pubsub.Subscription
	teh            *pubsub.TopicEventHandler
	validatorSet   *sets.ValidatorSet
	eventHandler   sets.PubSubEventHandler
	messageHandler sets.MessageHandler

	// statusCh is used to send a notification about a new message, it's
	// returned by the Transport.WaitFor function.
	statusCh chan transport.Status
}

func newSubscription(node *Node, topic string, typ transport.Message) (*Subscription, error) {
	var err error
	s := &Subscription{
		ctx:            node.ctx,
		validatorSet:   node.validatorSet,
		eventHandler:   node.pubSubEventHandlerSet,
		messageHandler: node.messageHandlerSet,
		statusCh:       make(chan transport.Status),
	}
	err = node.pubSub.RegisterTopicValidator(topic, s.validator(topic, typ))
	if err != nil {
		return nil, err
	}
	s.topic, err = node.PubSub().Join(topic)
	if err != nil {
		return nil, err
	}
	s.sub, err = s.topic.Subscribe()
	if err != nil {
		return nil, err
	}
	s.teh, err = s.topic.EventHandler()
	if err != nil {
		return nil, err
	}
	s.eventLoop()
	return s, err
}

func (s Subscription) Publish(message transport.Message) error {
	b, err := message.Marshall()
	if err != nil {
		return err
	}
	if b == nil {
		return ErrNilMessage
	}
	s.messageHandler.Published(s.topic.String(), b, message)
	return s.topic.Publish(s.ctx, b)
}

func (s Subscription) Next() chan transport.Status {
	go func() {
		var msg transport.Message
		psMsg, err := s.sub.Next(s.ctx)
		if psMsg != nil && err == nil {
			msg = psMsg.ValidatorData.(transport.Message)
		}
		s.statusCh <- transport.Status{
			Message: msg,
			Error:   err,
		}
	}()
	return s.statusCh
}

func (s Subscription) validator(topic string, typ transport.Message) pubsub.ValidatorEx {
	// Validator actually have two roles in the libp2p: it unmarshalls messages
	// and then validates them. Unmarshalled message is stored in the
	// ValidatorData field which was created for this purpose:
	// https://github.com/libp2p/go-libp2p-pubsub/pull/231
	r := reflect.TypeOf(typ).Elem()
	return func(ctx context.Context, id peer.ID, psMsg *pubsub.Message) pubsub.ValidationResult {
		msg := reflect.New(r).Interface().(transport.Message)
		err := msg.Unmarshall(psMsg.Data)
		if err != nil {
			s.messageHandler.Broken(topic, psMsg, err)
			return pubsub.ValidationReject
		}
		psMsg.ValidatorData = msg
		vr := s.validatorSet.Validator(topic)(ctx, id, psMsg)
		s.messageHandler.Received(topic, psMsg, vr)
		return vr
	}
}

func (s Subscription) eventLoop() {
	go func() {
		for {
			pe, err := s.teh.NextPeerEvent(s.ctx)
			if err != nil {
				// The only time when an error may be returned here is
				// when the subscription is canceled.
				return
			}
			s.eventHandler.Handle(s.topic.String(), pe)
		}
	}()
}

func (s Subscription) close() error {
	s.teh.Cancel()
	s.sub.Cancel()
	return s.topic.Close()
}
