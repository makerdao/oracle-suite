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
	"github.com/libp2p/go-libp2p-core/network"

	"github.com/makerdao/gofer/pkg/log"
	"github.com/makerdao/gofer/pkg/transport/p2p/sets"
)

type node interface {
	AddNotifee(notifees ...network.Notifiee)
	AddEventHandler(eventHandler ...sets.EventHandler)
	AddMessageHandler(messageHandlers ...sets.MessageHandler)
}

// Register registers p2p.Node extensions which will print additional
// debug logs.
func Register(node node, l log.Logger) {
	node.AddNotifee(&notifee{log: l})
	node.AddEventHandler(&eventHandler{log: l})
	node.AddMessageHandler(&messageHandler{log: l})
}