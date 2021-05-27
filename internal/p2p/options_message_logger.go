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
	pubsub "github.com/libp2p/go-libp2p-pubsub"

	"github.com/makerdao/oracle-suite/pkg/log"
	"github.com/makerdao/oracle-suite/pkg/transport"
)

// MessageLogger logs published and received messages.
func MessageLogger() Options {
	return func(n *Node) error {
		mlh := &messageLoggerHandler{node: n}
		n.AddMessageHandler(mlh)
		return nil
	}
}

type messageLoggerHandler struct {
	node *Node
}

func (m *messageLoggerHandler) Published(topic string, raw []byte, _ transport.Message) {
	m.node.log.
		WithFields(log.Fields{
			"topic":   topic,
			"message": string(raw),
		}).
		Debug("Published a new message")
}

func (m *messageLoggerHandler) Received(topic string, msg *pubsub.Message, _ pubsub.ValidationResult) {
	addrs := m.node.Host().Peerstore().PeerInfo(msg.ReceivedFrom).Addrs
	m.node.log.
		WithFields(log.Fields{
			"topic":              topic,
			"message":            string(msg.Data),
			"peerID":             msg.GetFrom().String(),
			"receivedFromPeerID": msg.ReceivedFrom.String(),
			"receivedFromAddrs":  addrs,
		}).
		Debug("Received a new message")
}

func (m *messageLoggerHandler) Broken(topic string, msg *pubsub.Message, err error) {
	addrs := m.node.Host().Peerstore().PeerInfo(msg.ReceivedFrom).Addrs
	m.node.log.
		WithError(err).
		WithFields(log.Fields{
			"topic":              topic,
			"peerID":             msg.GetFrom().String(),
			"receivedFromPeerID": msg.ReceivedFrom.String(),
			"receivedFromAddrs":  addrs,
		}).
		Debug("Unable to unmarshall received message")
}
