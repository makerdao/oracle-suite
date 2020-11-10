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
		Mask: net.CIDRMask(32, 32),
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
