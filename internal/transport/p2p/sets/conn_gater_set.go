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

package sets

import (
	"github.com/libp2p/go-libp2p-core/connmgr"
	"github.com/libp2p/go-libp2p-core/control"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
)

// ConnGaterSet implements the connmgr.ConnectionGater and allow to aggregate
// multiple instances of this interface.
type ConnGaterSet struct {
	connGaters []connmgr.ConnectionGater
}

// NewConnGaterSet creates new instance of the ConnGaterSet.
func NewConnGaterSet() *ConnGaterSet {
	return &ConnGaterSet{}
}

// Add adds new connmgr.ConnectionGater to the set.
func (n *ConnGaterSet) Add(connGaters ...connmgr.ConnectionGater) {
	n.connGaters = append(n.connGaters, connGaters...)
}

// InterceptAddrDial implements the connmgr.ConnectionGater interface.
func (f *ConnGaterSet) InterceptAddrDial(id peer.ID, addr multiaddr.Multiaddr) bool {
	for _, connGater := range f.connGaters {
		if !connGater.InterceptAddrDial(id, addr) {
			return false
		}
	}
	return true
}

// InterceptPeerDial implements the connmgr.ConnectionGater interface.
func (f *ConnGaterSet) InterceptPeerDial(id peer.ID) bool {
	for _, connGater := range f.connGaters {
		if !connGater.InterceptPeerDial(id) {
			return false
		}
	}
	return true
}

// InterceptAccept implements the connmgr.ConnectionGater interface.
func (f *ConnGaterSet) InterceptAccept(network network.ConnMultiaddrs) bool {
	for _, connGater := range f.connGaters {
		if !connGater.InterceptAccept(network) {
			return false
		}
	}
	return true
}

// InterceptSecured implements the connmgr.ConnectionGater interface.
func (f *ConnGaterSet) InterceptSecured(dir network.Direction, id peer.ID, network network.ConnMultiaddrs) bool {
	for _, connGater := range f.connGaters {
		if !connGater.InterceptSecured(dir, id, network) {
			return false
		}
	}
	return true
}

// InterceptUpgraded implements the connmgr.ConnectionGater interface.
func (f *ConnGaterSet) InterceptUpgraded(conn network.Conn) (bool, control.DisconnectReason) {
	for _, connGater := range f.connGaters {
		if allow, reason := connGater.InterceptUpgraded(conn); !allow {
			return allow, reason
		}
	}
	return true, 0
}
