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
