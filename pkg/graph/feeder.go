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
	"time"

	"github.com/hashicorp/go-multierror"

	"github.com/makerdao/gofer/pkg/origins"
)

type Feedable interface {
	// OriginPair returns the origin and pair which are acceptable for
	// this Node.
	OriginPair() OriginPair
	// Ingest sets the Tick for this Node. It may return error if
	// the OriginTick contains incompatible origin or pair.
	Ingest(tick OriginTick) error
	// MinTTL is the amount of time during which the Tick shouldn't be updated.
	MinTTL() time.Duration
	// MaxTTL is the maximum amount of time during which the Tick can be used.
	// After that time, the Tick method will return a OriginTick with
	// a TickTTLExpiredErr error.
	MaxTTL() time.Duration
	// Expired returns true if the Tick is expired. This is based on the MaxTTL
	// value.
	Expired() bool
	// Tick returns the Tick assigned in the Ingest method. If the Tick is
	// expired then a TickTTLExpiredErr error will be set in
	// the OriginTick.Error field.
	Tick() OriginTick
}

// Feeder sets ticks from origins to the Feedable nodes.
type Feeder struct {
	set *origins.Set
}

// NewFeeder creates new Feeder instance.
func NewFeeder(set *origins.Set) *Feeder {
	return &Feeder{set: set}
}

// Feed sets Ticks to Feedable nodes. This method takes list of root Nodes
// and sets Ticks to all of their children that implements the Feedable interface.
//
// This method may return an error with a list of problems during fetching, but
// despite this there may be enough data to calculate prices. To check that,
// invoke the Tick() method on the root node and check if there is an error
// in AggregatorTick.Error field.
func (i *Feeder) Feed(nodes ...Node) error {
	return i.fetchTicks(i.findFeedableNodes(nodes))
}

// findFeedableNodes returns a list of children nodes from given root nodes
// which implements Feedable interface.
func (i *Feeder) findFeedableNodes(nodes []Node) []Feedable {
	tn := time.Now()

	var feedables []Feedable
	Walk(func(node Node) {
		if feedable, ok := node.(Feedable); ok {
			tr := tn.Add(-1 * feedable.MinTTL())
			if feedable.Tick().Timestamp.Before(tr) {
				feedables = append(feedables, feedable)
			}
		}
	}, nodes...)

	return feedables
}

func (i *Feeder) fetchTicks(nodes []Feedable) error {
	var err error

	// originPair is used as a key in a map to easily find
	// Feedable nodes for given origin and pair
	type originPair struct {
		origin string
		pair   origins.Pair
	}

	nodesMap := map[originPair][]Feedable{}
	pairsMap := map[string][]origins.Pair{}

	for _, node := range nodes {
		originPair := originPair{
			origin: node.OriginPair().Origin,
			pair: origins.Pair{
				Base:  node.OriginPair().Pair.Base,
				Quote: node.OriginPair().Pair.Quote,
			},
		}

		nodesMap[originPair] = appendNodeIfUnique(
			nodesMap[originPair],
			node,
		)

		pairsMap[originPair.origin] = appendPairIfUnique(
			pairsMap[originPair.origin],
			originPair.pair,
		)
	}

	for origin, frs := range i.set.Fetch(pairsMap) {
		for _, fr := range frs {
			originPair := originPair{
				origin: origin,
				pair:   fr.Tick.Pair,
			}

			for _, feedable := range nodesMap[originPair] {
				tick := mapOriginResult(origin, fr)

				// If there was an error during fetching a Tick but previous Tick is still
				// not expired, do not try to override it:
				if tick.Error != nil && !feedable.Expired() {
					err = multierror.Append(err, tick.Error)
				} else {
					iErr := feedable.Ingest(tick)
					if iErr != nil {
						err = multierror.Append(err, feedable.Ingest(tick))
					}
				}
			}
		}
	}

	return err
}

func appendPairIfUnique(pairs []origins.Pair, pair origins.Pair) []origins.Pair {
	exists := false
	for _, p := range pairs {
		if p.Equal(pair) {
			exists = true
			break
		}
	}
	if !exists {
		pairs = append(pairs, pair)
	}

	return pairs
}

func appendNodeIfUnique(nodes []Feedable, node Feedable) []Feedable {
	exists := false
	for _, n := range nodes {
		if n == node {
			exists = true
			break
		}
	}
	if !exists {
		nodes = append(nodes, node)
	}

	return nodes
}

func mapOriginResult(origin string, fr origins.FetchResult) OriginTick {
	return OriginTick{
		Tick: Tick{
			Pair: Pair{
				Base:  fr.Tick.Pair.Base,
				Quote: fr.Tick.Pair.Quote,
			},
			Price:     fr.Tick.Price,
			Bid:       fr.Tick.Bid,
			Ask:       fr.Tick.Ask,
			Volume24h: fr.Tick.Volume24h,
			Timestamp: fr.Tick.Timestamp,
		},
		Origin: origin,
		Error:  fr.Error,
	}
}
