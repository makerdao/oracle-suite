package sets

import (
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/multiformats/go-multiaddr"
)

// NotifeeSet implements the network.Notifiee and allow to aggregate
// multiple instances of this interface.
type NotifeeSet struct {
	notifees []network.Notifiee
}

// NewNotifeeSet creates new instance of the NotifeeSet.
func NewNotifeeSet() *NotifeeSet {
	return &NotifeeSet{}
}

// Add adds new network.Notifiee to the set.
func (n *NotifeeSet) Add(notifees ...network.Notifiee) {
	n.notifees = append(n.notifees, notifees...)
}

// Listen implements the network.Notifiee interface.
func (n *NotifeeSet) Listen(network network.Network, maddr multiaddr.Multiaddr) {
	for _, notifee := range n.notifees {
		notifee.Listen(network, maddr)
	}
}

// ListenClose implements the network.Notifiee interface.
func (n *NotifeeSet) ListenClose(network network.Network, maddr multiaddr.Multiaddr) {
	for _, notifee := range n.notifees {
		notifee.ListenClose(network, maddr)
	}
}

// Connected implements the network.Notifiee interface.
func (n *NotifeeSet) Connected(network network.Network, conn network.Conn) {
	for _, notifee := range n.notifees {
		notifee.Connected(network, conn)
	}
}

// Disconnected implements the network.Notifiee interface.
func (n *NotifeeSet) Disconnected(network network.Network, conn network.Conn) {
	for _, notifee := range n.notifees {
		notifee.Disconnected(network, conn)
	}
}

// OpenedStream implements the network.Notifiee interface.
func (n *NotifeeSet) OpenedStream(network network.Network, stream network.Stream) {
	for _, notifee := range n.notifees {
		notifee.OpenedStream(network, stream)
	}
}

// ClosedStream implements the network.Notifiee interface.
func (n *NotifeeSet) ClosedStream(network network.Network, stream network.Stream) {
	for _, notifee := range n.notifees {
		notifee.ClosedStream(network, stream)
	}
}
