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

package graph

import (
	"fmt"
	"time"

	"github.com/hashicorp/go-multierror"
)

type IngestedIncompatiblePairErr struct {
	Given    Pair
	Expected Pair
}

func (e IngestedIncompatiblePairErr) Error() string {
	return fmt.Sprintf(
		"a tick with different pair ignested to the OriginNode, %s given but %s was expected",
		e.Given,
		e.Expected,
	)
}

type IngestedIncompatibleOriginErr struct {
	Given    string
	Expected string
}

func (e IngestedIncompatibleOriginErr) Error() string {
	return fmt.Sprintf(
		"a tick from different origin ignested to the OriginNode, %s given but %s was expected",
		e.Given,
		e.Expected,
	)
}

type TickTTLExpiredErr struct {
	Tick OriginTick
	TTL  time.Duration
}

func (e TickTTLExpiredErr) Error() string {
	return fmt.Sprintf(
		"the tick TTL for the pair %s expired",
		e.Tick.Pair,
	)
}

// OriginNode contains a Tick fetched directly from an origin.
type OriginNode struct {
	originPair OriginPair
	tick       OriginTick
	minTTL     time.Duration
	maxTTL     time.Duration
}

func NewOriginNode(originPair OriginPair, minTTL time.Duration, maxTTL time.Duration) *OriginNode {
	return &OriginNode{
		originPair: originPair,
		minTTL:     minTTL,
		maxTTL:     maxTTL,
	}
}

// OriginPair implements the Feedable interface.
func (n *OriginNode) OriginPair() OriginPair {
	return n.originPair
}

// Ingest implements Feedable interface.
func (n *OriginNode) Ingest(tick OriginTick) error {
	var err error
	if !tick.Pair.Equal(n.originPair.Pair) {
		err = multierror.Append(err, IngestedIncompatiblePairErr{
			Given:    tick.Pair,
			Expected: n.originPair.Pair,
		})
	}

	if tick.Origin != n.originPair.Origin {
		err = multierror.Append(err, IngestedIncompatibleOriginErr{
			Given:    tick.Origin,
			Expected: n.originPair.Origin,
		})
	}

	if err == nil {
		n.tick = tick
	}

	return err
}

// MinTTL implements the Feedable interface.
func (n *OriginNode) MinTTL() time.Duration {
	return n.minTTL
}

// MaxTTL implements the Feedable interface.
func (n *OriginNode) MaxTTL() time.Duration {
	return n.maxTTL
}

// Expired implements the Feedable interface.
func (n OriginNode) Expired() bool {
	return n.tick.Timestamp.Before(time.Now().Add(-1 * n.MaxTTL()))
}

// Tick implements the Feedable interface.
func (n *OriginNode) Tick() OriginTick {
	if n.tick.Error == nil {
		if n.Expired() {
			n.tick.Error = TickTTLExpiredErr{
				Tick: n.tick,
				TTL:  n.maxTTL,
			}
		}
	}

	return n.tick
}

// Children implements the Node interface.
func (n OriginNode) Children() []Node {
	return []Node{}
}
