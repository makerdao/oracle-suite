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

package local

import (
	"context"
	"errors"
	"reflect"

	"github.com/makerdao/oracle-suite/pkg/transport"
)

var ErrNotSubscribed = errors.New("topic is not subscribed")

// Local is a simple implementation of the transport.Transport interface
// using local channels.
type Local struct {
	ctx    context.Context
	doneCh chan struct{}

	buffer int
	subs   map[string]subscription
}

type subscription struct {
	typ    reflect.Type
	msgs   chan []byte
	status chan transport.ReceivedMessage
}

// New returns a new instance of the Local structure.
func New(ctx context.Context, buffer int, subs map[string]transport.Message) *Local {
	l := &Local{
		ctx:    ctx,
		doneCh: make(chan struct{}),
		buffer: buffer,
		subs:   make(map[string]subscription),
	}
	for topic, typ := range subs {
		l.subs[topic] = subscription{
			typ:    reflect.TypeOf(typ).Elem(),
			msgs:   make(chan []byte, l.buffer),
			status: make(chan transport.ReceivedMessage),
		}
	}
	return l
}

// Start implements the transport.Transport interface.
func (l *Local) Start() error {
	// Handle context cancellation:
	go func() {
		defer func() { l.doneCh <- struct{}{} }()
		<-l.ctx.Done()
		for _, sub := range l.subs {
			close(sub.status)
		}
		l.subs = make(map[string]subscription)
	}()
	return nil
}

// Wait implements the transport.Transport interface.
func (l *Local) Wait() {
	<-l.doneCh
}

// Broadcast implements the transport.Transport interface.
func (l *Local) Broadcast(topic string, message transport.Message) error {
	if sub, ok := l.subs[topic]; ok {
		b, err := message.Marshall()
		if err != nil {
			return err
		}
		sub.msgs <- b
		return nil
	}
	return ErrNotSubscribed
}

// WaitFor implements the transport.Transport interface.
func (l *Local) WaitFor(topic string) chan transport.ReceivedMessage {
	if sub, ok := l.subs[topic]; ok {
		go func() {
			select {
			case <-l.ctx.Done():
				return
			case msg := <-sub.msgs:
				message := reflect.New(sub.typ).Interface().(transport.Message)
				sub.status <- transport.ReceivedMessage{
					Message: message,
					Error:   message.Unmarshall(msg),
				}
			}
		}()
		return sub.status
	}
	return nil
}
