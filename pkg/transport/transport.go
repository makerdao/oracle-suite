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

type ReceivedMessage struct {
	// Message contains the message content. It is nil when the Error field
	// is not nil.
	Message Message
	// Data contains an optional data associated with the message. A type of
	// the data is different depending on a transport implementation.
	Data interface{}
	// Error contains an optional error returned by a transport.
	Error error
}

type Message interface {
	Marshall() ([]byte, error)
	Unmarshall([]byte) error
}

// Transport is the interface for different implementations of a
// publish–subscribe messaging solutions for the Oracle network.
type Transport interface {
	Broadcast(topic string, message Message) error
	// WaitFor returns a channel which will be blocked until message for given
	// topic arrives. Note, that only messages for subscribed topics will
	// be supported by this method, for unsubscribed topic nil will be
	// returned. In case of an error, error will be returned in a Status
	// structure.
	WaitFor(topic string) chan ReceivedMessage
	// Start starts listening for messages.
	Start() error
	// Stop stops listening for messages.
	Stop() error
}
