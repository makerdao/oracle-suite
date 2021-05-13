package node

import (
	"net"

	"github.com/libp2p/go-libp2p-core/control"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"

	"github.com/makerdao/oracle-suite/pkg/log"
)

func Denylist(addrs []multiaddr.Multiaddr) Options {
	return func(n *Node) error {
		cg := &denylistConnGater{log: n.log}
		n.AddConnectionGater(cg)
		for _, maddr := range addrs {
			multiaddr.ForEach(maddr, func(c multiaddr.Component) bool {
				switch c.Protocol().Code {
				case multiaddr.P_IP4, multiaddr.P_IP6:
					cg.BlockIP(net.ParseIP(c.String()))
				case multiaddr.P_P2P:
					pid, err := peer.IDFromBytes(c.RawValue())
					if err != nil {
						return true
					}
					cg.BlockPID(pid)
				}
				return true
			})
		}
		return nil
	}
}

type denylistConnGater struct {
	filters multiaddr.Filters
	pids    []peer.ID
	log     log.Logger
}

// BlockPID blocks connections from given peer ID.
func (f *denylistConnGater) BlockPID(pid peer.ID) {
	f.pids = append(f.pids, pid)
}

// BlockIP blocks connections from given IP address.
func (f *denylistConnGater) BlockIP(ip net.IP) {
	f.filters.AddFilter(net.IPNet{
		IP:   ip,
		Mask: net.CIDRMask(len(ip)*8, len(ip)*8),
	}, multiaddr.ActionDeny)
}

// InterceptAddrDial implements the connmgr.ConnectionGater interface.
func (f *denylistConnGater) InterceptAddrDial(pid peer.ID, addr multiaddr.Multiaddr) bool {
	if f.filters.AddrBlocked(addr) {
		return false
	}
	for _, p := range f.pids {
		if p == pid {
			return false
		}
	}
	return true
}

// InterceptPeerDial implements the connmgr.ConnectionGater interface.
func (f *denylistConnGater) InterceptPeerDial(peer.ID) bool {
	return true
}

// InterceptAccept implements the connmgr.ConnectionGater interface.
func (f *denylistConnGater) InterceptAccept(network.ConnMultiaddrs) bool {
	return true
}

// InterceptSecured implements the connmgr.ConnectionGater interface.
func (f *denylistConnGater) InterceptSecured(network.Direction, peer.ID, network.ConnMultiaddrs) bool {
	return true
}

// InterceptUpgraded implements the connmgr.ConnectionGater interface.
func (f *denylistConnGater) InterceptUpgraded(network.Conn) (bool, control.DisconnectReason) {
	return true, 0
}
