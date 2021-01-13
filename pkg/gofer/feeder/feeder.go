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

package feeder

import (
	"time"

	"github.com/hashicorp/go-multierror"

	"github.com/makerdao/gofer/pkg/gofer/graph"
	"github.com/makerdao/gofer/pkg/gofer/origins"
	"github.com/makerdao/gofer/pkg/log"
)

const LoggerTag = "FEEDER"

type Feedable interface {
	// OriginPair returns the origin and pair which are acceptable for
	// this Node.
	OriginPair() graph.OriginPair
	// Ingest sets the Tick for this Node. It may return error if
	// the OriginTick contains incompatible origin or pair.
	Ingest(tick graph.OriginTick) error
	// MinTTL is the amount of time during which the Tick shouldn't be updated.
	MinTTL() time.Duration
	// MaxTTL is the maximum amount of time during which the Tick can be used.
	// After that time, the Tick method will return a OriginTick with
	// a ErrTickTTLExpired error.
	MaxTTL() time.Duration
	// Expired returns true if the Tick is expired. This is based on the MaxTTL
	// value.
	Expired() bool
	// Tick returns the Tick assigned in the Ingest method. If the Tick is
	// expired then a ErrTickTTLExpired error will be set in
	// the OriginTick.Error field.
	Tick() graph.OriginTick
}

// Feeder sets ticks from origins to the Feedable nodes.
type Feeder struct {
	set    *origins.Set
	log    log.Logger
	doneCh chan bool
}

// NewFeeder creates new Feeder instance.
func NewFeeder(set *origins.Set, log log.Logger) *Feeder {
	return &Feeder{
		set:    set,
		log:    log.WithField("tag", LoggerTag),
		doneCh: make(chan bool),
	}
}

// Feed sets Ticks to Feedable nodes. This method takes list of root Nodes
// and sets Ticks to all of their children that implement the Feedable interface.
//
// This method may return an error with a list of problems during fetching, but
// despite this there may be enough data to calculate prices. To check that,
// invoke the Tick() method on the root node and check if there is an error
// in AggregatorTick.Error field.
func (f *Feeder) Feed(nodes []graph.Node) error {
	return f.fetchTicksAndFeedThemToFeedableNodes(f.findFeedableNodes(nodes))
}

func (f *Feeder) Start(nodes []graph.Node) error {
	f.log.Infof("Starting")

	err := f.Feed(nodes)
	if err != nil {
		f.log.WithError(err).Info("Unable to feed some nodes")
	}

	ticker := time.NewTicker(getMinTTL(nodes))
	go func() {
		for {
			select {
			case <-f.doneCh:
				ticker.Stop()
				return
			case <-ticker.C:
				err := f.Feed(nodes)
				if err != nil {
					f.log.WithError(err).Info("Unable to feed some nodes")
				}
			}
		}
	}()

	return nil
}

func (f *Feeder) Stop() {
	defer f.log.Infof("Stopped")

	f.doneCh <- true
}

// findFeedableNodes returns a list of children nodes from given root nodes
// which implement Feedable interface.
func (f *Feeder) findFeedableNodes(nodes []graph.Node) []Feedable {
	tn := time.Now()

	var feedables []Feedable
	graph.Walk(func(node graph.Node) {
		if feedable, ok := node.(Feedable); ok {
			tr := tn.Add(-1 * feedable.MinTTL())
			if feedable.Tick().Timestamp.Before(tr) {
				feedables = append(feedables, feedable)
			}
		}
	}, nodes...)

	return feedables
}

func (f *Feeder) fetchTicksAndFeedThemToFeedableNodes(nodes []Feedable) error {
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

	for origin, frs := range f.set.Fetch(pairsMap) {
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
				} else if iErr := feedable.Ingest(tick); iErr != nil {
					err = multierror.Append(iErr, feedable.Ingest(tick))
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

func mapOriginResult(origin string, fr origins.FetchResult) graph.OriginTick {
	return graph.OriginTick{
		Tick: graph.Tick{
			Pair: graph.Pair{
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

func getMinTTL(nodes []graph.Node) time.Duration {
	minTTL := time.Duration(0)
	graph.Walk(func(node graph.Node) {
		if feedable, ok := node.(Feedable); ok {
			if minTTL == 0 || feedable.MinTTL() < minTTL {
				minTTL = feedable.MinTTL()
			}
		}
	}, nodes...)

	if minTTL < time.Second {
		return time.Second
	}

	return minTTL
}
