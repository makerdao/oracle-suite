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

package transport

type Status struct {
	Message interface{}
	Error   error
}

type Message interface {
	Marshall() ([]byte, error)
	Unmarshall([]byte) error
}

type Transport interface {
	// Subscribe starts subscribing for messages with given topic. The second
	// argument is a type of a message given as a nil pointer,
	// eg: (*Message)(nil).
	Subscribe(topic string, typ Message) error
	// Unsubscribe stops subscribing for messages with given topic.
	Unsubscribe(topic string) error
	// Broadcast sends a message with given topic to the network. To send
	// a message, you must first subscribe appropriate topic.
	Broadcast(topic string, message Message) error
	// WaitFor returns a channel which will be blocked until message for given
	// topic arrives. Then message will be unmarshalled to type given in the
	// Subscribe method. Note, that only messages for subscribed topics will
	// be supported by this method, for unsubscribed topic nil will be
	// returned. In case of an error, error will be returned in a Status
	// structure.
	WaitFor(topic string) chan Status
	// Close closes connection.
	Close() error
}
