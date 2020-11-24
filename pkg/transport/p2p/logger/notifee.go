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
	"github.com/multiformats/go-multiaddr"

	"github.com/makerdao/gofer/pkg/log"
)

type notifee struct {
	log log.Logger
}

func (n *notifee) Listen(network.Network, multiaddr.Multiaddr) {}

func (n *notifee) ListenClose(network.Network, multiaddr.Multiaddr) {}

func (n *notifee) Connected(network network.Network, conn network.Conn) {
	n.log.
		WithFields(log.Fields{"ip": conn.LocalMultiaddr().String()}).
		Debug("Connected to a host")
}

func (n *notifee) Disconnected(network network.Network, conn network.Conn) {
	n.log.
		WithFields(log.Fields{"ip": conn.LocalMultiaddr().String()}).
		Debug("Disconnected from a host")
}

func (n *notifee) OpenedStream(network.Network, network.Stream) {}

func (n *notifee) ClosedStream(network.Network, network.Stream) {}
