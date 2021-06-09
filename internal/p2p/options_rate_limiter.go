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
	"sync"
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"golang.org/x/time/rate"

	"github.com/makerdao/oracle-suite/pkg/log"
)

// RateLimiterConfig is a configuration for the RateLimiter option.
type RateLimiterConfig struct {
	// PeerBytesPerSecond is the maximum number of bytes/s that can be
	// received from a single peer.
	PeerBytesPerSecond float64
	// GlobalBurst is a burst value in bytes applied for a messages received
	// from a singe peer.
	PeerBurst int
	// GlobalBytesPerSecond is the maximum number of bytes/s that can be
	// received from the network. Messages rejected by peer limiters are not
	// counted.
	GlobalBytesPerSecond float64
	// GlobalBurst is a burst value in bytes applied for a messages received
	// from the network. Messages rejected by peer limiters are not counted.
	GlobalBurst int
}

type rateLimiter struct {
	mu sync.Mutex

	peerBtsPerSec float64
	peerBurstSize int
	peerLimiters  map[peer.ID]*peerLimiter

	// globalLimiter is used for all messages from any peer.
	globalLimiter *rate.Limiter
	// gcTTL is a time since last message after which peer will be removed
	// by the gc method.
	gcTTL time.Duration
}

type peerLimiter struct {
	limiter *rate.Limiter
	lastMsg time.Time // lastMsg is a time since last message.
}

// peerLimiter creates or returns previously created limiter for a given peer.
func (p *rateLimiter) peerLimiter(id peer.ID) *peerLimiter {
	p.mu.Lock()
	defer p.mu.Unlock()
	if _, ok := p.peerLimiters[id]; !ok {
		p.peerLimiters[id] = &peerLimiter{
			limiter: rate.NewLimiter(rate.Limit(p.peerBtsPerSec), p.peerBurstSize),
			lastMsg: time.Now(),
		}
	}
	return p.peerLimiters[id]
}

// allow checks if a message of a given size can be received from given peer.
func (p *rateLimiter) allow(id peer.ID, msgSize int) bool {
	prl := p.peerLimiter(id)
	prl.lastMsg = time.Now()
	// It is important, that global limiter cannot be called if peer limiter
	// returns false. Otherwise, a misbehaving peer may exhaust the limits
	// for the entire network.
	return prl.limiter.AllowN(prl.lastMsg, msgSize) && p.globalLimiter.AllowN(time.Now(), msgSize)
}

// gc removes inactive peers.
func (p *rateLimiter) gc() {
	p.mu.Lock()
	defer p.mu.Unlock()
	for id, pl := range p.peerLimiters {
		if time.Since(pl.lastMsg) >= p.gcTTL {
			delete(p.peerLimiters, id)
		}
	}
}

// RateLimiter limits number of bytes which is allowed to receive from
// the network using token bucket algorithm:
// https://en.wikipedia.org/wiki/Token_bucket
func RateLimiter(cfg RateLimiterConfig) Options {
	return func(n *Node) error {
		rl := &rateLimiter{
			peerBtsPerSec: cfg.PeerBytesPerSecond,
			peerBurstSize: cfg.PeerBurst,
			globalLimiter: rate.NewLimiter(rate.Limit(cfg.GlobalBytesPerSecond), cfg.GlobalBurst),
			peerLimiters:  make(map[peer.ID]*peerLimiter),
			// gcTTL is a time required to fill empty token bucket, then multiplied by two:
			gcTTL: time.Second * time.Duration(float64(cfg.PeerBurst)/cfg.PeerBytesPerSecond) * 2,
		}
		n.AddValidator(func(ctx context.Context, topic string, id peer.ID, msg *pubsub.Message) pubsub.ValidationResult {
			// TODO: Should we do another check for msg.ReceivedFrom?
			if rl.allow(msg.GetFrom(), len(msg.Data)) {
				return pubsub.ValidationAccept
			}
			n.log.
				WithFields(log.Fields{
					"topic":              topic,
					"peerID":             msg.GetFrom().String(),
					"receivedFromPeerID": msg.ReceivedFrom.String(),
				}).
				Debug("The message was rejected due to rate limiting")
			return pubsub.ValidationIgnore
		})
		go func() {
			t := time.NewTimer(time.Minute)
			defer t.Stop()
			for {
				select {
				case <-n.ctx.Done():
					return
				case <-t.C:
					rl.gc()
				}
			}
		}()
		return nil
	}
}
