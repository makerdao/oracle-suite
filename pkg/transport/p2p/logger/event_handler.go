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
)

type eventHandler struct {
	log log.Logger
}

func (e eventHandler) Handle(topic string, event pubsub.PeerEvent) {
	switch event.Type {
	case pubsub.PeerJoin:
		e.log.
			WithFields(log.Fields{"id": event.Peer.Pretty(), "topic": topic}).
			Debug("Connected to a peer")
	case pubsub.PeerLeave:
		e.log.
			WithFields(log.Fields{"id": event.Peer.Pretty(), "topic": topic}).
			Debug("Disconnected from a peer")
	}
}
