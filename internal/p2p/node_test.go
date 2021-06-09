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

package p2p

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/makerdao/oracle-suite/pkg/transport"
)

const defaultTimeout = 10 * time.Second

func TestNode_AddrNotLeaking(t *testing.T) {
	// This test checks that the addresses of nodes that do not use discovery
	// are not revealed to other nodes in the network.
	//
	// Topology:
	//   n1 <--[discovery]--> n2 <--[direct]--> n3
	//
	// - n1 node should only be connected to n2 node (through discovery)
	// - n2 node should only be connected to n1 node (through discovery) and n2 (through direct connection)
	// - n3 node should only be connected to n2 node (through direct connection)
	// - the n1 node's address must not be exposed to n3 node

	peers, err := GetPeerInfo(3)
	require.NoError(t, err)

	n1, err := NewNode(
		context.Background(),
		PeerPrivKey(peers[0].PrivKey),
		ListenAddrs(peers[0].ListenAddrs),
		Discovery(nil),
	)
	require.NoError(t, err)
	require.NoError(t, n1.Start())
	defer n1.Stop()

	n2, err := NewNode(
		context.Background(),
		PeerPrivKey(peers[1].PrivKey),
		ListenAddrs(peers[1].ListenAddrs),
		DirectPeers(peers[2].PeerAddrs),
		Discovery(peers[0].PeerAddrs),
	)
	require.NoError(t, err)
	require.NoError(t, n2.Start())
	defer n2.Stop()

	n3, err := NewNode(
		context.Background(),
		PeerPrivKey(peers[2].PrivKey),
		ListenAddrs(peers[2].ListenAddrs),
		DirectPeers(peers[1].PeerAddrs),
	)
	require.NoError(t, err)
	require.NoError(t, n3.Start())
	defer n3.Stop()

	require.NoError(t, n1.Subscribe("test", (*Message)(nil)))
	require.NoError(t, n2.Subscribe("test", (*Message)(nil)))
	require.NoError(t, n3.Subscribe("test", (*Message)(nil)))

	WaitFor(t, func() bool {
		lp := n1.PubSub().ListPeers("test")
		return len(lp) == 1 && ContainsPeerID(lp, peers[1].ID)
	}, defaultTimeout)
	WaitFor(t, func() bool {
		lp := n2.PubSub().ListPeers("test")
		return len(lp) == 2 && ContainsPeerID(lp, peers[0].ID) && ContainsPeerID(lp, peers[2].ID)
	}, defaultTimeout)
	WaitFor(t, func() bool {
		lp := n3.PubSub().ListPeers("test")
		return len(lp) == 1 && ContainsPeerID(lp, peers[1].ID)
	}, defaultTimeout)
}

func TestNode_MessagePropagation(t *testing.T) {
	// This test checks if messages are propagated between peers correctly.
	//
	// Topology:
	//   n1 <--[manual connection]--> n2

	peers, err := GetPeerInfo(2)
	require.NoError(t, err)

	n1, err := NewNode(
		context.Background(),
		PeerPrivKey(peers[0].PrivKey),
		ListenAddrs(peers[0].ListenAddrs),
	)
	require.NoError(t, err)
	require.NoError(t, n1.Start())
	defer n1.Stop()

	n2, err := NewNode(
		context.Background(),
		PeerPrivKey(peers[1].PrivKey),
		ListenAddrs(peers[1].ListenAddrs),
	)
	require.NoError(t, err)
	require.NoError(t, n2.Start())
	defer n2.Stop()

	require.NoError(t, n1.Connect(peers[1].PeerAddrs[0]))
	require.NoError(t, n1.Subscribe("test", (*Message)(nil)))
	require.NoError(t, n2.Subscribe("test", (*Message)(nil)))

	WaitFor(t, func() bool {
		return len(n1.PubSub().ListPeers("test")) > 0 && len(n2.PubSub().ListPeers("test")) > 0
	}, defaultTimeout)

	s1, err := n1.Subscription("test")
	require.NoError(t, err)
	s2, err := n2.Subscription("test")
	require.NoError(t, err)

	err = s1.Publish(NewMessage("makerdao"))
	assert.NoError(t, err)

	// Message should be received on both nodes:
	WaitForMessage(t, s1.Next(), NewMessage("makerdao"), defaultTimeout)
	WaitForMessage(t, s2.Next(), NewMessage("makerdao"), defaultTimeout)
}

func TestNode_ConnectionLimit(t *testing.T) {
	peers, err := GetPeerInfo(5)
	require.NoError(t, err)

	n, err := NewNode(
		context.Background(),
		PeerPrivKey(peers[0].PrivKey),
		ListenAddrs(peers[0].ListenAddrs),
		ConnectionLimit(1, 1, 0),
	)
	require.NoError(t, err)
	require.NoError(t, n.Start())
	defer n.Stop()

	for i := 2; i < len(peers); i++ {
		n, err := NewNode(
			context.Background(),
			PeerPrivKey(peers[i].PrivKey),
			ListenAddrs(peers[i].ListenAddrs),
			Discovery(nil),
		)
		require.NoError(t, err)
		require.NoError(t, n.Start())
		defer n.Stop()

		require.NoError(t, n.Connect(peers[0].PeerAddrs[0]))
	}

	n.Host().ConnManager().TrimOpenConns(context.Background())
	time.Sleep(time.Second)

	conns := 0
	for _, p := range n.Host().Peerstore().Peers() {
		if n.Host().Network().Connectedness(p) == network.Connected {
			conns++
		}
	}

	assert.Equal(t, conns, 1)
}

// Message is the simplest implementation of the transport.Message interface.
type Message []byte

// NewMessage Returns a new Message.
func NewMessage(msg string) *Message {
	b := Message(msg)
	return &b
}

func (m *Message) String() string {
	if m == nil {
		return ""
	}
	return string(*m)
}

func (m *Message) Equal(msg *Message) bool {
	return bytes.Equal(*m, *msg)
}

func (m *Message) Marshall() ([]byte, error) {
	return *m, nil
}

func (m *Message) Unmarshall(bytes []byte) error {
	*m = bytes
	return nil
}

type PeerInfo struct {
	ID          peer.ID
	PrivKey     crypto.PrivKey
	ListenAddrs []multiaddr.Multiaddr
	PeerAddrs   []multiaddr.Multiaddr
}

// GetPeerInfo returns n PeerInfo structs which can be used to generate
// random test nodes.
func GetPeerInfo(n int) ([]PeerInfo, error) {
	ps, err := GetFreePorts(n)
	if err != nil {
		return nil, err
	}
	var pi []PeerInfo
	for i := 0; i < n; i++ {
		rr := rand.Reader
		sk, _, err := crypto.GenerateEd25519Key(rr)
		if err != nil {
			return nil, err
		}
		id, err := peer.IDFromPrivateKey(sk)
		if err != nil {
			return nil, err
		}
		pi = append(pi, PeerInfo{
			ListenAddrs: []multiaddr.Multiaddr{multiaddr.StringCast(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", ps[i]))},
			PeerAddrs:   []multiaddr.Multiaddr{multiaddr.StringCast(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d/p2p/%s", ps[i], id.Pretty()))},
			PrivKey:     sk,
			ID:          id,
		})
	}
	return pi, nil
}

// GetFreePorts returns n random ports available to use.
func GetFreePorts(n int) ([]int, error) {
	var ports []int
	for i := 0; i < n; i++ {
		addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
		if err != nil {
			return nil, err
		}
		l, err := net.ListenTCP("tcp", addr)
		if err != nil {
			return nil, err
		}
		defer l.Close()
		ports = append(ports, l.Addr().(*net.TCPAddr).Port)
	}
	return ports, nil
}

func WaitFor(t *testing.T, cond func() bool, timeout time.Duration) {
	s := time.Now()
	for !cond() {
		if time.Since(s) >= timeout {
			assert.Fail(t, "timeout")
			return
		}
		time.Sleep(time.Second)
	}
}

func WaitForMessage(t *testing.T, stat chan transport.ReceivedMessage, expected *Message, timeout time.Duration) {
	to := time.After(timeout)
	select {
	case received := <-stat:
		require.NoError(t, received.Error, "subscription returned an error")
		receivedBts, err := received.Message.Marshall()
		if err != nil {
			assert.NoError(t, err, "unable to unmarshall received message")
		}
		expectedBts, err := expected.Marshall()
		if err != nil {
			assert.NoError(t, err, "unable to unmarshall expected message")
		}
		assert.Equal(t, expectedBts, receivedBts)
	case <-to:
		assert.Fail(t, "timeout")
		return
	}
}

func CountMessages(sub *Subscription, duration time.Duration) chan map[peer.ID]int {
	ch := make(chan map[peer.ID]int, 0)
	go func() {
		count := map[peer.ID]int{}
		defer func() { ch <- count }()
		for {
			select {
			case <-time.After(duration):
				return
			case msg, ok := <-sub.Next():
				if !ok {
					return
				}
				id := msg.Data.(*pubsub.Message).GetFrom()
				if _, ok := count[id]; !ok {
					count[id] = 0
				}
				count[id]++
			}
		}
	}()
	return ch
}

func ContainsPeerID(ids []peer.ID, id peer.ID) bool {
	for _, i := range ids {
		if i == id {
			return true
		}
	}
	return false
}
