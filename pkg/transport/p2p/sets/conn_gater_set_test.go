package sets

import (
	"testing"

	"github.com/libp2p/go-libp2p-core/control"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/assert"
)

type testConnGater struct {
	Allow  bool
	Reason control.DisconnectReason
}

func (t *testConnGater) InterceptPeerDial(_ peer.ID) bool {
	return t.Allow
}

func (t *testConnGater) InterceptAddrDial(_ peer.ID, _ multiaddr.Multiaddr) bool {
	return t.Allow
}

func (t *testConnGater) InterceptAccept(_ network.ConnMultiaddrs) bool {
	return t.Allow
}

func (t *testConnGater) InterceptSecured(_ network.Direction, _ peer.ID, _ network.ConnMultiaddrs) bool {
	return t.Allow
}

func (t *testConnGater) InterceptUpgraded(_ network.Conn) (bool, control.DisconnectReason) {
	return t.Allow, t.Reason
}

func TestConnGaterSet_InterceptAccept(t *testing.T) {
	a := &testConnGater{}
	b := &testConnGater{}

	gns := NewConnGaterSet()
	gns.Add(a)
	gns.Add(b)

	a.Allow = true
	b.Allow = true
	assert.True(t, gns.InterceptAccept((network.ConnMultiaddrs)(nil)))

	a.Allow = false
	b.Allow = false
	assert.False(t, gns.InterceptAccept((network.ConnMultiaddrs)(nil)))

	a.Allow = false
	b.Allow = true
	assert.False(t, gns.InterceptAccept((network.ConnMultiaddrs)(nil)))

	a.Allow = true
	b.Allow = false
	assert.False(t, gns.InterceptAccept((network.ConnMultiaddrs)(nil)))
}

func TestConnGaterSet_InterceptAddrDial(t *testing.T) {
	a := &testConnGater{}
	b := &testConnGater{}

	gns := NewConnGaterSet()
	gns.Add(a)
	gns.Add(b)

	a.Allow = true
	b.Allow = true
	assert.True(t, gns.InterceptAddrDial("", (multiaddr.Multiaddr)(nil)))

	a.Allow = false
	b.Allow = false
	assert.False(t, gns.InterceptAddrDial("", (multiaddr.Multiaddr)(nil)))

	a.Allow = false
	b.Allow = true
	assert.False(t, gns.InterceptAddrDial("", (multiaddr.Multiaddr)(nil)))

	a.Allow = true
	b.Allow = false
	assert.False(t, gns.InterceptAddrDial("", (multiaddr.Multiaddr)(nil)))
}

func TestConnGaterSet_InterceptPeerDial(t *testing.T) {
	a := &testConnGater{}
	b := &testConnGater{}

	gns := NewConnGaterSet()
	gns.Add(a)
	gns.Add(b)

	a.Allow = true
	b.Allow = true
	assert.True(t, gns.InterceptPeerDial(""))

	a.Allow = false
	b.Allow = false
	assert.False(t, gns.InterceptPeerDial(""))

	a.Allow = false
	b.Allow = true
	assert.False(t, gns.InterceptPeerDial(""))

	a.Allow = true
	b.Allow = false
	assert.False(t, gns.InterceptPeerDial(""))
}

func TestConnGaterSet_InterceptSecured(t *testing.T) {
	a := &testConnGater{}
	b := &testConnGater{}

	gns := NewConnGaterSet()
	gns.Add(a)
	gns.Add(b)

	a.Allow = true
	b.Allow = true
	assert.True(t, gns.InterceptSecured(network.DirUnknown, "", (network.ConnMultiaddrs)(nil)))

	a.Allow = false
	b.Allow = false
	assert.False(t, gns.InterceptSecured(network.DirUnknown, "", (network.ConnMultiaddrs)(nil)))

	a.Allow = false
	b.Allow = true
	assert.False(t, gns.InterceptSecured(network.DirUnknown, "", (network.ConnMultiaddrs)(nil)))

	a.Allow = true
	b.Allow = false
	assert.False(t, gns.InterceptSecured(network.DirUnknown, "", (network.ConnMultiaddrs)(nil)))
}

func TestConnGaterSet_InterceptUpgraded(t *testing.T) {
	a := &testConnGater{}
	b := &testConnGater{}

	gns := NewConnGaterSet()
	gns.Add(a)
	gns.Add(b)

	var allow bool
	var reason control.DisconnectReason

	a.Allow = true
	b.Allow = true
	allow, reason = gns.InterceptUpgraded((network.Conn)(nil))
	assert.True(t, allow)

	a.Allow = false
	b.Allow = false
	a.Reason = 1
	a.Reason = 2
	allow, reason = gns.InterceptUpgraded((network.Conn)(nil))
	assert.False(t, allow)
	assert.Equal(t, a.Reason, reason)

	a.Allow = false
	b.Allow = true
	allow, reason = gns.InterceptUpgraded((network.Conn)(nil))
	assert.False(t, allow)
	assert.Equal(t, a.Reason, reason)

	a.Allow = true
	b.Allow = false
	allow, reason = gns.InterceptUpgraded((network.Conn)(nil))
	assert.False(t, allow)
	assert.Equal(t, b.Reason, reason)
}
