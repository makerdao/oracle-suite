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

	"github.com/makerdao/oracle-suite/pkg/transport"
)

var ErrAlreadySubscribed = errors.New("topic is already subscribed")
var ErrNotSubscribed = errors.New("topic is not subscribed")

// Local is a simple implementation of the transport.Transport interface
// using local channels.
type Local struct {
	buffer int
	subs   map[string]subscription
}

type subscription struct {
	msgs   chan []byte
	status chan transport.Status
	doneCh chan struct{}
}

// New returns a new instance of the Local structure.
func New(buffer int) *Local {
	return &Local{
		buffer: buffer,
		subs:   make(map[string]subscription),
	}
}

// Subscribe implements the transport.Transport interface.
func (l *Local) Subscribe(topic string) error {
	if _, ok := l.subs[topic]; ok {
		return ErrAlreadySubscribed
	}
	l.subs[topic] = subscription{
		msgs:   make(chan []byte, l.buffer),
		status: make(chan transport.Status),
		doneCh: make(chan struct{}),
	}
	return nil
}

// Unsubscribe implements the transport.Transport interface.
func (l *Local) Unsubscribe(topic string) error {
	if sub, ok := l.subs[topic]; ok {
		close(sub.doneCh)
		close(sub.status)
		delete(l.subs, topic)
		return nil
	}
	return ErrNotSubscribed
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
func (l *Local) WaitFor(topic string, message transport.Message) chan transport.Status {
	if sub, ok := l.subs[topic]; ok {
		go func() {
			select {
			case <-sub.doneCh:
				return
			case msg := <-sub.msgs:
				sub.status <- transport.Status{
					Message: message,
					Error:   message.Unmarshall(msg),
				}
			}
		}()
		return sub.status
	}
	return nil
}

// Close implements the transport.Transport interface.
func (l *Local) Close() error {
	for _, sub := range l.subs {
		close(sub.doneCh)
		close(sub.status)
	}
	l.subs = make(map[string]subscription)
	return nil
}
