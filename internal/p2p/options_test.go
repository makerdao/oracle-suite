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
	"context"
	"crypto/rand"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNode_PeerPrivKey(t *testing.T) {
	// This tests checks the PeerPrivKey option.

	sk, _, _ := crypto.GenerateRSAKeyPair(2048, rand.Reader)

	n, err := NewNode(
		context.Background(),
		PeerPrivKey(sk),
	)
	require.NoError(t, err)
	require.NoError(t, n.Start())
	defer n.Stop()

	id, _ := peer.IDFromPrivateKey(sk)
	assert.Equal(t, id, n.Host().ID())
}

func TestNode_MessagePrivKey(t *testing.T) {
	// This tests checks the MessagePrivKey option.

	sk, _, _ := crypto.GenerateRSAKeyPair(2048, rand.Reader)

	n, err := NewNode(
		context.Background(),
		MessagePrivKey(sk),
	)
	require.NoError(t, err)
	require.NoError(t, n.Start())
	defer n.Stop()

	require.NoError(t, n.Subscribe("test", (*Message)(nil)))
	s, err := n.Subscription("test")
	require.NoError(t, err)

	err = s.Publish(NewMessage("makerdao"))
	require.NoError(t, err)

	// The public key used to sign the message should be derived from the key
	// passed to the MessagePrivKey function:
	id, _ := peer.IDFromPrivateKey(sk)
	msg := <-s.Next()
	require.NoError(t, msg.Error)
	assert.Equal(t, id, msg.Data.(*pubsub.Message).GetFrom())
	// The public key extracted form a message must be different
	// than peer's public key:
	assert.NotEqual(t, n.Host().ID(), msg.Data.(*pubsub.Message).GetFrom())
}

func TestNode_Discovery(t *testing.T) {
	// This test checks whether all nodes in the network can discover each
	// other when Discovery option is used.
	//
	// Topology:
	//   n1 <--[discovery]--> n2 <--[discovery]--> n3

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
		Discovery(peers[0].PeerAddrs),
	)
	require.NoError(t, err)
	require.NoError(t, n2.Start())
	defer n2.Stop()

	n3, err := NewNode(
		context.Background(),
		PeerPrivKey(peers[2].PrivKey),
		ListenAddrs(peers[2].ListenAddrs),
		Discovery(peers[0].PeerAddrs),
	)
	require.NoError(t, err)
	require.NoError(t, n3.Start())
	defer n3.Stop()

	require.NoError(t, n1.Subscribe("test", (*Message)(nil)))
	require.NoError(t, n2.Subscribe("test", (*Message)(nil)))
	require.NoError(t, n3.Subscribe("test", (*Message)(nil)))

	// Every peer should see two other peers:
	WaitFor(t, func() bool {
		lp := n1.PubSub().ListPeers("test")
		return ContainsPeerID(lp, peers[1].ID) && ContainsPeerID(lp, peers[2].ID)
	}, defaultTimeout)
	WaitFor(t, func() bool {
		lp := n2.PubSub().ListPeers("test")
		return ContainsPeerID(lp, peers[0].ID) && ContainsPeerID(lp, peers[2].ID)
	}, defaultTimeout)
	WaitFor(t, func() bool {
		lp := n3.PubSub().ListPeers("test")
		return ContainsPeerID(lp, peers[0].ID) && ContainsPeerID(lp, peers[1].ID)
	}, defaultTimeout)
}

func TestNode_ConnectionLimit(t *testing.T) {
	// This test checks whether the connection number is properly limited when
	// the ConnectionLimit option is used.

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

func TestNode_DirectPeers(t *testing.T) {
	// This test checks whether the direct connection between peers configured
	// using the DirectPeers option is always maintained.

	peers, err := GetPeerInfo(5)
	require.NoError(t, err)

	n1, err := NewNode(
		context.Background(),
		PeerPrivKey(peers[0].PrivKey),
		ListenAddrs(peers[0].ListenAddrs),
		ConnectionLimit(1, 1, 0),
		DirectPeers(peers[1].PeerAddrs),
	)
	require.NoError(t, err)
	require.NoError(t, n1.Start())
	defer n1.Stop()

	n2, err := NewNode(
		context.Background(),
		PeerPrivKey(peers[1].PrivKey),
		ListenAddrs(peers[1].ListenAddrs),
		DirectPeers(peers[0].PeerAddrs),
	)
	require.NoError(t, err)
	require.NoError(t, n2.Start())
	defer n2.Stop()

	for i := 2; i < len(peers); i++ {
		n, err := NewNode(
			context.Background(),
			PeerPrivKey(peers[i].PrivKey),
			ListenAddrs(peers[i].ListenAddrs),
		)
		require.NoError(t, err)
		require.NoError(t, n.Start())
		defer n.Stop()

		require.NoError(t, n.Connect(peers[0].PeerAddrs[0]))
		// Connection with tagged hosts are less likely to be dropped.
		// By tagging them we can be sure it wasn't coincidence that
		// the connection to n2 host is maintained after call to
		// the TrimOpenConns method.
		n1.Host().ConnManager().TagPeer(n.Host().ID(), "test", 1)
	}

	// The connection between n1 and n2 nodes should be persisted even
	// with a connection limit:
	n1.Host().ConnManager().TrimOpenConns(context.Background())
	time.Sleep(time.Second)
	assert.Equal(t, network.Connected, n1.Host().Network().Connectedness(n2.Host().ID()))
}
