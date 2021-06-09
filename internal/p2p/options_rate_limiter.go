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
)

type RateLimiterConfig struct {
	// GlobalBytesPerSecond is a maximum number of allowed bytes/s to be
	// received from the network.
	GlobalBytesPerSecond float64
	// PeerBytesPerSecond is a maximum number of allowed bytes/s to be
	// received from a single peer.
	PeerBytesPerSecond float64
	// GlobalBurst is a burst value in bytes applied for a
	// messages received from the network.
	GlobalBurst int
	// GlobalBurst is a burst value in bytes applied for a
	// messages received from a singe peer.
	PeerBurst int
}

type rateLimiter struct {
	mu sync.Mutex

	peerBtsPerSec float64
	peerBurstSize int
	globalLimiter *rate.Limiter
	peerLimiters  map[peer.ID]*peerLimiter
}

type peerLimiter struct {
	limiter *rate.Limiter
	lastMsg time.Time
}

func (p *rateLimiter) peerLimiter(id peer.ID) *peerLimiter {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, ok := p.peerLimiters[id]; !ok {
		p.peerLimiters[id] = &peerLimiter{
			limiter: rate.NewLimiter(rate.Limit(p.peerBtsPerSec), p.peerBurstSize),
		}
	}

	return p.peerLimiters[id]
}

func (p *rateLimiter) allowPeer(id peer.ID, msgSize int) bool {
	prl := p.peerLimiter(id)
	prl.lastMsg = time.Now()

	return prl.limiter.AllowN(prl.lastMsg, msgSize)
}

func (p *rateLimiter) allowGlobal(msgSize int) bool {
	return p.globalLimiter.AllowN(time.Now(), msgSize)
}

func (p *rateLimiter) allow(id peer.ID, msgSize int) bool {
	return p.allowPeer(id, msgSize) && p.allowGlobal(msgSize)
}

func (p *rateLimiter) gc(ttl time.Duration) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for id, pl := range p.peerLimiters {
		if time.Since(pl.lastMsg) >= ttl {
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
			peerLimiters:  make(map[peer.ID]*peerLimiter, 0),
		}
		go func() {
			ttl := time.Second * time.Duration(float64(cfg.PeerBurst)/cfg.PeerBytesPerSecond)
			t := time.NewTimer(time.Minute)
			defer t.Stop()
			for {
				select {
				case <-t.C:
					rl.gc(ttl)
				case <-n.ctx.Done():
					return
				}
			}
		}()
		n.AddValidator(func(ctx context.Context, topic string, id peer.ID, msg *pubsub.Message) pubsub.ValidationResult {
			if rl.allow(id, len(msg.Data)) {
				return pubsub.ValidationAccept
			}
			return pubsub.ValidationIgnore
		})
		return nil
	}
}
