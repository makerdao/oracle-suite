package p2p

import (
	"net"

	"github.com/libp2p/go-libp2p-core/control"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"

	"github.com/makerdao/gofer/internal/log"
)

type connGater struct {
	bannedAddrs multiaddr.Filters
	log         log.Logger
}

func newConnGater(l log.Logger) *connGater {
	return &connGater{
		log: l,
	}
}

func (f *connGater) banIP(ip net.IP) {
	f.bannedAddrs.AddFilter(net.IPNet{
		IP:   ip,
		Mask: net.CIDRMask(32, 32),
	}, multiaddr.ActionDeny)
}

func (f *connGater) InterceptAddrDial(id peer.ID, addr multiaddr.Multiaddr) bool {
	if f.bannedAddrs.AddrBlocked(addr) {
		f.log.
			WithFields(log.Fields{"id": id.Pretty(), "addr": addr}).
			Debug("Unable to connect to banned peer")

		return false
	}

	return true
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
