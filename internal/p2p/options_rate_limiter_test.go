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
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNode_RateLimiter_PeerLimit(t *testing.T) {
	// This test checks if peer limits works correctly. The limit is set to
	// 128 bytes/s, we will try to send two messages of 128 bytes each.
	// The second one must be rejected because it will exceed the 128 bytes/s
	// limit. Then we wait one second and try to send another 128 byte message.
	// This time the message should be accepted.

	peers, err := GetPeerInfo(2)
	require.NoError(t, err)

	n1, err := NewNode(
		context.Background(),
		PeerPrivKey(peers[0].PrivKey),
		ListenAddrs(peers[0].ListenAddrs),
		Discovery(nil),
		RateLimiter(RateLimiterConfig{
			GlobalBytesPerSecond: 1024,
			PeerBytesPerSecond:   128,
			GlobalBurst:          1024,
			PeerBurst:            128,
		}),
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

	require.NoError(t, n1.Subscribe("test", (*Message)(nil)))
	require.NoError(t, n2.Subscribe("test", (*Message)(nil)))

	s1, err := n1.Subscription("test")
	require.NoError(t, err)
	s2, err := n2.Subscription("test")
	require.NoError(t, err)

	// Wait for the peers to connect to each other:
	WaitFor(t, func() bool {
		return len(n1.PubSub().ListPeers("test")) > 0 && len(n2.PubSub().ListPeers("test")) > 0
	}, defaultTimeout)

	// Send messages:
	msgsCh := CountMessages(s1, 2*time.Second)
	msg := NewMessage(strings.Repeat("a", 128))
	require.NoError(t, s2.Publish(msg))
	require.NoError(t, s2.Publish(msg)) // exceeds limit
	time.Sleep(1 * time.Second)
	require.NoError(t, s2.Publish(msg))

	// Only two messages should arrive, rest messages exceed the peer limit:
	assert.Equal(t, 2, (<-msgsCh)[n2.Host().ID()])
}

func TestNode_RateLimiter_PeerBurst(t *testing.T) {
	// This test checks if data burst for a peer works correctly. The value for
	// the data limit is smaller than the message size. We will try to send two
	// messages. The first one should be accepted because its size is within the
	// burst limit. The second one should be rejected because it exceeds the
	// burst limit.

	peers, err := GetPeerInfo(2)
	require.NoError(t, err)

	n1, err := NewNode(
		context.Background(),
		PeerPrivKey(peers[0].PrivKey),
		ListenAddrs(peers[0].ListenAddrs),
		Discovery(nil),
		RateLimiter(RateLimiterConfig{
			GlobalBytesPerSecond: 1,
			PeerBytesPerSecond:   1,
			GlobalBurst:          3 * 1024,
			PeerBurst:            1024,
		}),
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

	require.NoError(t, n1.Subscribe("test", (*Message)(nil)))
	require.NoError(t, n2.Subscribe("test", (*Message)(nil)))

	s1, err := n1.Subscription("test")
	require.NoError(t, err)
	s2, err := n2.Subscription("test")
	require.NoError(t, err)

	// Wait for the peers to connect to each other:
	WaitFor(t, func() bool {
		return len(n1.PubSub().ListPeers("test")) > 0 && len(n2.PubSub().ListPeers("test")) > 0
	}, defaultTimeout)

	// Send messages:
	msgsCh := CountMessages(s1, 2*time.Second)
	msg := NewMessage(strings.Repeat("a", 1024))
	require.NoError(t, s2.Publish(msg))
	require.NoError(t, s2.Publish(msg))

	// Only one message should arrive, second one exceeds the peer limit:
	assert.Equal(t, 1, (<-msgsCh)[n2.Host().ID()])
}

func TestNode_RateLimiter_GlobalBurst(t *testing.T) {
	// This test checks if global data burst works correctly. The value for
	// the global data limit is smaller than the message size. We will try to
	// send two messages. The first one should be accepted because its size is
	// within the burst limit. The second one should be rejected because it
	// exceeds the burst limit.

	peers, err := GetPeerInfo(2)
	require.NoError(t, err)

	n1, err := NewNode(
		context.Background(),
		PeerPrivKey(peers[0].PrivKey),
		ListenAddrs(peers[0].ListenAddrs),
		Discovery(nil),
		RateLimiter(RateLimiterConfig{
			GlobalBytesPerSecond: 1,
			PeerBytesPerSecond:   1,
			GlobalBurst:          1024,
			PeerBurst:            3 * 1024,
		}),
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

	require.NoError(t, n1.Subscribe("test", (*Message)(nil)))
	require.NoError(t, n2.Subscribe("test", (*Message)(nil)))

	s1, err := n1.Subscription("test")
	require.NoError(t, err)
	s2, err := n2.Subscription("test")
	require.NoError(t, err)

	// Wait for the peers to connect to each other:
	WaitFor(t, func() bool {
		return len(n1.PubSub().ListPeers("test")) > 0 && len(n2.PubSub().ListPeers("test")) > 0
	}, defaultTimeout)

	// Send messages:
	msgsCh := CountMessages(s1, 2*time.Second)
	msg := NewMessage(strings.Repeat("a", 1024))
	require.NoError(t, s2.Publish(msg))
	require.NoError(t, s2.Publish(msg))

	// Only one message should arrive, second one exceeds the peer limit:
	assert.Equal(t, 1, (<-msgsCh)[n2.Host().ID()])
}
