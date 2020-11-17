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

package banner

import (
	"net"

	"github.com/libp2p/go-libp2p-core/control"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
)

type connGater struct {
	bannedAddrs multiaddr.Filters
}

func (f *connGater) BanIP(ip net.IP) {
	f.bannedAddrs.AddFilter(net.IPNet{
		IP:   ip,
		Mask: net.CIDRMask(len(ip)*8, len(ip)*8),
	}, multiaddr.ActionDeny)
}

func (f *connGater) InterceptAddrDial(id peer.ID, addr multiaddr.Multiaddr) bool {
	return !f.bannedAddrs.AddrBlocked(addr)
}

func (f *connGater) InterceptPeerDial(peer.ID) bool {
	return true
}

func (f *connGater) InterceptAccept(network.ConnMultiaddrs) bool {
	return true
}

func (f *connGater) InterceptSecured(network.Direction, peer.ID, network.ConnMultiaddrs) bool {
	return true
}

func (f *connGater) InterceptUpgraded(network.Conn) (bool, control.DisconnectReason) {
	return true, 0
}
