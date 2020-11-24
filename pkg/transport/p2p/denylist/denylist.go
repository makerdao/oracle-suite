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

package denylist

import (
	"net"

	"github.com/libp2p/go-libp2p-core/connmgr"
	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/multiformats/go-multiaddr"
)

type node interface {
	AddConnectionGater(connGaters ...connmgr.ConnectionGater)
	PubSub() *pubsub.PubSub
}

// Register registers p2p.Node extensions required by the Denylist and returns
// its instance.
func Register(node node) *Denylist {
	denylist := &Denylist{
		connGater: &connGater{},
		pubSub:    node.PubSub(),
	}
	node.AddConnectionGater(denylist.connGater)
	return denylist
}

// Denylist allow to block connections to nodes using their IP or ID.
type Denylist struct {
	connGater *connGater
	pubSub    *pubsub.PubSub
}

// Deny blocks neither by an IP or a peer ID. If provided multiaddress contains
// an IP and a peer ID, both will be blocked separately.
func (b *Denylist) Deny(maddr multiaddr.Multiaddr) error {
	multiaddr.ForEach(maddr, func(c multiaddr.Component) bool {
		switch c.Protocol().Code {
		case multiaddr.P_IP4, multiaddr.P_IP6:
			b.denyIP(net.ParseIP(c.String()))
		case multiaddr.P_P2P:
			id, err := peer.IDFromBytes(c.RawValue())
			if err != nil {
				return true
			}
			b.denyID(id)
		}
		return true
	})
	return nil
}

func (b *Denylist) denyIP(ip net.IP) {
	b.connGater.BlockIP(ip)
}

func (b *Denylist) denyID(id peer.ID) {
	b.pubSub.BlacklistPeer(id)
}
