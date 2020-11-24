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

package logger

import (
	pubsub "github.com/libp2p/go-libp2p-pubsub"

	"github.com/makerdao/gofer/pkg/log"
	"github.com/makerdao/gofer/pkg/transport"
)

type messageHandler struct {
	log log.Logger
}

func (m *messageHandler) Published(topic string, raw []byte, message transport.Message) {
	m.log.
		WithFields(log.Fields{"topic": topic, "message": string(raw)}).
		Debug("Published a new message")
}

func (m *messageHandler) Received(topic string, raw *pubsub.Message, message transport.Message) {
	m.log.
		WithFields(log.Fields{"topic": topic, "message": string(raw.Data), "peerID": raw.ReceivedFrom}).
		Debug("Received a new message")
}
