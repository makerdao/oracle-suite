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

	"github.com/libp2p/go-libp2p-core/connmgr"
	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/multiformats/go-multiaddr"
)

type node interface {
	AddConnectionGater(connGaters ...connmgr.ConnectionGater)
	PubSub() *pubsub.PubSub
}

// Register registers p2p.Node extensions required by the Banner and returns
// its instance.
func Register(node node) *Banner {
	banner := &Banner{
		connGater: &connGater{},
		pubSub:    node.PubSub(),
	}
	node.AddConnectionGater(banner.connGater)
	return banner
}

// Banner allow to ban connections to nodes using their IP or ID. All connections
// to banned nodes will be blocked.
type Banner struct {
	connGater *connGater
	pubSub    *pubsub.PubSub
}

// Ban bans neither by an IP or a peer ID. If provided multiaddress contains
// an IP and a peer ID, both will be blocked separately.
func (b *Banner) Ban(maddr multiaddr.Multiaddr) error {
	multiaddr.ForEach(maddr, func(c multiaddr.Component) bool {
		switch c.Protocol().Code {
		case multiaddr.P_IP4, multiaddr.P_IP6:
			b.BanIP(net.ParseIP(c.String()))
		case multiaddr.P_P2P:
			id, err := peer.IDFromBytes(c.RawValue())
			if err != nil {
				return true
			}
			b.BanPeer(id)
		}
		return true
	})

	return nil
}

// BanIP bans given IP address.
func (b *Banner) BanIP(ip net.IP) {
	b.connGater.BanIP(ip)
}

// BanIP bans given peer ID.
func (b *Banner) BanPeer(id peer.ID) {
	b.pubSub.BlacklistPeer(id)
}
