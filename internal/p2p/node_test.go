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
	//   n0 <--[discovery]--> n1 <--[direct]--> n2
	//
	// - n0 node should only be connected to n1 node (through discovery)
	// - n1 node should only be connected to n0 node (through discovery) and n1 (through direct connection)
	// - n2 node should only be connected to n1 node (through direct connection)
	// - the n0 node's address must not be exposed to n2 node

	peers, err := getPeerInfo(3)
	require.NoError(t, err)

	n0, err := NewNode(
		context.Background(),
		PeerPrivKey(peers[0].PrivKey),
		ListenAddrs(peers[0].ListenAddrs),
		Discovery(nil),
	)
	require.NoError(t, err)
	require.NoError(t, n0.Start())
	defer n0.Stop()

	n1, err := NewNode(
		context.Background(),
		PeerPrivKey(peers[1].PrivKey),
		ListenAddrs(peers[1].ListenAddrs),
		DirectPeers(peers[2].PeerAddrs),
		Discovery(peers[0].PeerAddrs),
	)
	require.NoError(t, err)
	require.NoError(t, n1.Start())
	defer n1.Stop()

	n2, err := NewNode(
		context.Background(),
		PeerPrivKey(peers[2].PrivKey),
		ListenAddrs(peers[2].ListenAddrs),
		DirectPeers(peers[1].PeerAddrs),
	)
	require.NoError(t, err)
	require.NoError(t, n2.Start())
	defer n2.Stop()

	require.NoError(t, n0.Subscribe("test", (*message)(nil)))
	require.NoError(t, n1.Subscribe("test", (*message)(nil)))
	require.NoError(t, n2.Subscribe("test", (*message)(nil)))

	waitFor(t, func() bool {
		lp := n0.PubSub().ListPeers("test")
		return len(lp) == 1 && containsPeerID(lp, peers[1].ID)
	})
	waitFor(t, func() bool {
		lp := n1.PubSub().ListPeers("test")
		return len(lp) == 2 && containsPeerID(lp, peers[0].ID) && containsPeerID(lp, peers[2].ID)
	})
	waitFor(t, func() bool {
		lp := n2.PubSub().ListPeers("test")
		return len(lp) == 1 && containsPeerID(lp, peers[1].ID)
	})
}

func TestNode_MessagePropagation(t *testing.T) {
	// This test checks if messages are propagated between peers correctly.
	//
	// Topology:
	//   n0 <--[manual connection]--> n1 <--[manual connection]--> n2

	peers, err := getPeerInfo(3)
	require.NoError(t, err)

	n0, err := NewNode(
		context.Background(),
		PeerPrivKey(peers[0].PrivKey),
		ListenAddrs(peers[0].ListenAddrs),
	)
	require.NoError(t, err)
	require.NoError(t, n0.Start())
	defer n0.Stop()

	n1, err := NewNode(
		context.Background(),
		PeerPrivKey(peers[1].PrivKey),
		ListenAddrs(peers[1].ListenAddrs),
	)
	require.NoError(t, err)
	require.NoError(t, n1.Start())
	defer n1.Stop()

	n2, err := NewNode(
		context.Background(),
		PeerPrivKey(peers[2].PrivKey),
		ListenAddrs(peers[2].ListenAddrs),
	)
	require.NoError(t, err)
	require.NoError(t, n2.Start())
	defer n1.Stop()

	require.NoError(t, n0.Connect(peers[1].PeerAddrs[0]))
	require.NoError(t, n1.Connect(peers[2].PeerAddrs[0]))
	require.NoError(t, n0.Subscribe("test", (*message)(nil)))
	require.NoError(t, n1.Subscribe("test", (*message)(nil)))
	require.NoError(t, n2.Subscribe("test", (*message)(nil)))

	// Wait for the peers to connect to each other:
	waitFor(t, func() bool {
		return len(n0.PubSub().ListPeers("test")) > 0 &&
			len(n1.PubSub().ListPeers("test")) > 0 &&
			len(n2.PubSub().ListPeers("test")) > 0
	})

	s1, err := n0.Subscription("test")
	require.NoError(t, err)
	s2, err := n1.Subscription("test")
	require.NoError(t, err)
	s3, err := n2.Subscription("test")
	require.NoError(t, err)

	err = s1.Publish(newMessage("makerdao"))
	assert.NoError(t, err)

	// Message should be received on both nodes:
	waitForMessage(t, s1.Next(), newMessage("makerdao"))
	waitForMessage(t, s2.Next(), newMessage("makerdao"))
	waitForMessage(t, s3.Next(), newMessage("makerdao"))
}

// message is the simplest implementation of the transport.Message interface.
type message []byte

// newMessage returns a new message.
func newMessage(msg string) *message {
	b := message(msg)
	return &b
}

func (m *message) String() string {
	if m == nil {
		return ""
	}
	return string(*m)
}

func (m *message) Equal(msg *message) bool {
	return bytes.Equal(*m, *msg)
}

func (m *message) Marshall() ([]byte, error) {
	return *m, nil
}

func (m *message) Unmarshall(bytes []byte) error {
	*m = bytes
	return nil
}

type peerInfo struct {
	ID          peer.ID
	PrivKey     crypto.PrivKey
	ListenAddrs []multiaddr.Multiaddr
	PeerAddrs   []multiaddr.Multiaddr
}

// getPeerInfo returns n peerInfo structs which can be used to generate
// random test nodes.
func getPeerInfo(n int) ([]peerInfo, error) {
	ps, err := getFreePorts(n)
	if err != nil {
		return nil, err
	}
	var pi []peerInfo
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
		pi = append(pi, peerInfo{
			ListenAddrs: []multiaddr.Multiaddr{multiaddr.StringCast(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", ps[i]))},
			PeerAddrs:   []multiaddr.Multiaddr{multiaddr.StringCast(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d/p2p/%s", ps[i], id.Pretty()))},
			PrivKey:     sk,
			ID:          id,
		})
	}
	return pi, nil
}

// getFreePorts returns n random ports available to use.
func getFreePorts(n int) ([]int, error) {
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

// waitFor waits until cond becomes true.
func waitFor(t *testing.T, cond func() bool) {
	s := time.Now()
	for !cond() {
		if time.Since(s) >= defaultTimeout {
			assert.Fail(t, "timeout")
			return
		}
		time.Sleep(time.Second)
	}
}

// waitForMessage waits for expected message.
func waitForMessage(t *testing.T, stat chan transport.ReceivedMessage, expected *message) {
	to := time.After(defaultTimeout)
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

// countMessages counts asynchronously received messages for specified time
// duration, then returns results in channel.
func countMessages(sub *Subscription, duration time.Duration) chan map[peer.ID]int {
	ch := make(chan map[peer.ID]int)
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

func containsPeerID(ids []peer.ID, id peer.ID) bool {
	for _, i := range ids {
		if i == id {
			return true
		}
	}
	return false
}
