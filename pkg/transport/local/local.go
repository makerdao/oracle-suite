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
	"errors"
	"reflect"

	"github.com/makerdao/oracle-suite/pkg/transport"
)

var ErrNotSubscribed = errors.New("topic is not subscribed")

// Local is a simple implementation of the transport.Transport interface
// using local channels.
type Local struct {
	buffer int
	subs   map[string]subscription
}

type subscription struct {
	typ    reflect.Type
	msgs   chan []byte
	status chan transport.ReceivedMessage
	doneCh chan struct{}
}

// New returns a new instance of the Local structure.
func New(buffer int, subs map[string]transport.Message) *Local {
	l := &Local{
		buffer: buffer,
		subs:   make(map[string]subscription),
	}
	for topic, typ := range subs {
		l.subs[topic] = subscription{
			typ:    reflect.TypeOf(typ).Elem(),
			msgs:   make(chan []byte, l.buffer),
			status: make(chan transport.ReceivedMessage),
			doneCh: make(chan struct{}),
		}
	}
	return l
}

// Start implements the transport.Transport interface.
func (l *Local) Start() error {
	return nil
}

// Stop implements the transport.Transport interface.
func (l *Local) Stop() error {
	for _, sub := range l.subs {
		close(sub.doneCh)
		close(sub.status)
	}
	l.subs = make(map[string]subscription)
	return nil
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
			case <-sub.doneCh:
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
