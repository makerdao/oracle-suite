package sets

import (
	"testing"

	"github.com/libp2p/go-libp2p-core/network"
	"github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/assert"
)

type testNotifee struct {
	ListenFn       func(network.Network, multiaddr.Multiaddr)
	ListenCloseFn  func(network.Network, multiaddr.Multiaddr)
	ConnectedFn    func(network.Network, network.Conn)
	DisconnectedFn func(network.Network, network.Conn)
	OpenedStreamFn func(network.Network, network.Stream)
	ClosedStreamFn func(network.Network, network.Stream)
}

func (t *testNotifee) Listen(network network.Network, multiaddr multiaddr.Multiaddr) {
	t.ListenFn(network, multiaddr)
}

func (t *testNotifee) ListenClose(network network.Network, multiaddr multiaddr.Multiaddr) {
	t.ListenCloseFn(network, multiaddr)
}

func (t *testNotifee) Connected(network network.Network, conn network.Conn) {
	t.ConnectedFn(network, conn)
}

func (t *testNotifee) Disconnected(network network.Network, conn network.Conn) {
	t.DisconnectedFn(network, conn)
}

func (t *testNotifee) OpenedStream(network network.Network, stream network.Stream) {
	t.OpenedStreamFn(network, stream)
}

func (t *testNotifee) ClosedStream(network network.Network, stream network.Stream) {
	t.ClosedStreamFn(network, stream)
}

func TestNotifeeSet_Connected(t *testing.T) {
	ns := NewNotifeeSet()

	n1 := &testNotifee{}
	n2 := &testNotifee{}
	ns.Add(n1, n2)

	calls := 0
	n1.ConnectedFn = func(n network.Network, c network.Conn) {
		calls++
	}
	n2.ConnectedFn = func(n network.Network, c network.Conn) {
		calls++
	}
	ns.Connected((network.Network)(nil), (network.Conn)(nil))

	assert.Equal(t, 2, calls)
}

func TestNotifeeSet_Disconnected(t *testing.T) {
	ns := NewNotifeeSet()

	n1 := &testNotifee{}
	n2 := &testNotifee{}
	ns.Add(n1, n2)

	calls := 0
	n1.DisconnectedFn = func(n network.Network, c network.Conn) {
		calls++
	}
	n2.DisconnectedFn = func(n network.Network, c network.Conn) {
		calls++
	}
	ns.Disconnected((network.Network)(nil), (network.Conn)(nil))

	assert.Equal(t, 2, calls)
}

func TestNotifeeSet_Listen(t *testing.T) {
	ns := NewNotifeeSet()

	n1 := &testNotifee{}
	n2 := &testNotifee{}
	ns.Add(n1, n2)

	calls := 0
	n1.ListenFn = func(n network.Network, m multiaddr.Multiaddr) {
		calls++
	}
	n2.ListenFn = func(n network.Network, m multiaddr.Multiaddr) {
		calls++
	}
	ns.Listen((network.Network)(nil), (multiaddr.Multiaddr)(nil))

	assert.Equal(t, 2, calls)
}

func TestNotifeeSet_ListenClose(t *testing.T) {
	ns := NewNotifeeSet()

	n1 := &testNotifee{}
	n2 := &testNotifee{}
	ns.Add(n1, n2)

	calls := 0
	n1.ListenCloseFn = func(n network.Network, m multiaddr.Multiaddr) {
		calls++
	}
	n2.ListenCloseFn = func(n network.Network, m multiaddr.Multiaddr) {
		calls++
	}
	ns.ListenClose((network.Network)(nil), (multiaddr.Multiaddr)(nil))

	assert.Equal(t, 2, calls)
}

func TestNotifeeSet_OpenedStream(t *testing.T) {
	ns := NewNotifeeSet()

	n1 := &testNotifee{}
	n2 := &testNotifee{}
	ns.Add(n1, n2)

	calls := 0
	n1.OpenedStreamFn = func(n network.Network, s network.Stream) {
		calls++
	}
	n2.OpenedStreamFn = func(n network.Network, s network.Stream) {
		calls++
	}
	ns.OpenedStream((network.Network)(nil), (network.Stream)(nil))

	assert.Equal(t, 2, calls)
}

func TestNotifeeSet_ClosedStream(t *testing.T) {
	ns := NewNotifeeSet()

	n1 := &testNotifee{}
	n2 := &testNotifee{}
	ns.Add(n1, n2)

	calls := 0
	n1.ClosedStreamFn = func(n network.Network, s network.Stream) {
		calls++
	}
	n2.ClosedStreamFn = func(n network.Network, s network.Stream) {
		calls++
	}
	ns.ClosedStream((network.Network)(nil), (network.Stream)(nil))

	assert.Equal(t, 2, calls)
}
