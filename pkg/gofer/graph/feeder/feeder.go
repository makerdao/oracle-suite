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

	"github.com/makerdao/oracle-suite/pkg/gofer"
	"github.com/makerdao/oracle-suite/pkg/gofer/graph/nodes"
	"github.com/makerdao/oracle-suite/pkg/gofer/origins"
	"github.com/makerdao/oracle-suite/pkg/log"
)

const LoggerTag = "FEEDER"

// Warnings contains a list of minor errors which occurred during fetching
// prices.
type Warnings struct {
	List []error
}

func (w Warnings) ToError() error {
	var err error
	for _, e := range w.List {
		err = multierror.Append(err, e)
	}
	return err
}

type Feedable interface {
	// OriginPair returns the origin and pair which are acceptable for
	// this Node.
	OriginPair() nodes.OriginPair
	// Ingest sets the Price for this Node. It may return error if
	// the OriginPrice contains incompatible origin or pair.
	Ingest(price nodes.OriginPrice) error
	// MinTTL is the amount of time during which the Price shouldn't be updated.
	MinTTL() time.Duration
	// MaxTTL is the maximum amount of time during which the Price can be used.
	// After that time, the Price method will return a OriginPrice with
	// a ErrPriceTTLExpired error.
	MaxTTL() time.Duration
	// Expired returns true if the Price is expired. This is based on the MaxTTL
	// value.
	Expired() bool
	// Price returns the Price assigned in the Ingest method. If the Price is
	// expired then a ErrPriceTTLExpired error will be set in
	// the OriginPrice.Error field.
	Price() nodes.OriginPrice
}

// Feeder sets prices from origins to the Feedable nodes.
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

// Feed sets Prices to Feedable nodes. This method takes list of root Models
// and sets Prices to all of their children that implement the Feedable interface.
func (f *Feeder) Feed(ns ...nodes.Node) Warnings {
	return f.fetchPricesAndFeedThemToFeedableNodes(f.findFeedableNodes(ns))
}

// Start starts a goroutine which updates prices as often as the lowest TTL is.
func (f *Feeder) Start(ns ...nodes.Node) error {
	f.log.Infof("Starting")

	warns := f.Feed(ns...)
	if len(warns.List) > 0 {
		f.log.WithError(warns.ToError()).Warn("Unable to feed some nodes")
	}

	ticker := time.NewTicker(getMinTTL(ns))
	go func() {
		for {
			select {
			case <-f.doneCh:
				ticker.Stop()
				return
			case <-ticker.C:
				warns := f.Feed(ns...)
				if len(warns.List) > 0 {
					f.log.WithError(warns.ToError()).Warn("Unable to feed some nodes")
				}
			}
		}
	}()

	return nil
}

// Stop stops a goroutine created by the Start method.
func (f *Feeder) Stop() {
	defer f.log.Infof("Stopped")

	f.doneCh <- true
}

// findFeedableNodes returns a list of children nodes from given root nodes
// which implement Feedable interface.
func (f *Feeder) findFeedableNodes(ns []nodes.Node) []Feedable {
	//tn := time.Now()

	var feedables []Feedable
	nodes.Walk(func(n nodes.Node) {
		if feedable, ok := n.(Feedable); ok {
			//if tn.Sub(feedable.Price().Time) > feedable.MinTTL() {
			feedables = append(feedables, feedable)
			//}
		}
	}, ns...)

	return feedables
}

func (f *Feeder) fetchPricesAndFeedThemToFeedableNodes(ns []Feedable) Warnings {
	var warns Warnings

	// originPair is used as a key in a map to easily find
	// Feedable nodes for given origin and pair
	type originPair struct {
		origin string
		pair   origins.Pair
	}

	nodesMap := map[originPair][]Feedable{}
	pairsMap := map[string][]origins.Pair{}

	for _, n := range ns {
		op := originPair{
			origin: n.OriginPair().Origin,
			pair: origins.Pair{
				Base:  n.OriginPair().Pair.Base,
				Quote: n.OriginPair().Pair.Quote,
			},
		}

		nodesMap[op] = appendNodeIfUnique(
			nodesMap[op],
			n,
		)

		pairsMap[op.origin] = appendPairIfUnique(
			pairsMap[op.origin],
			op.pair,
		)
	}

	for origin, frs := range f.set.Fetch(pairsMap) {
		for _, fr := range frs {
			op := originPair{
				origin: origin,
				pair:   fr.Price.Pair,
			}

			for _, feedable := range nodesMap[op] {
				price := mapOriginResult(origin, fr)

				// If there was an error during fetching a Price but previous Price is still
				// not expired, do not try to override it:
				if price.Error != nil && !feedable.Expired() {
					warns.List = append(warns.List, price.Error)
				} else if iErr := feedable.Ingest(price); iErr != nil {
					warns.List = append(warns.List, iErr)
				}
			}
		}
	}

	return warns
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

func appendNodeIfUnique(ns []Feedable, f Feedable) []Feedable {
	exists := false
	for _, n := range ns {
		if n == f {
			exists = true
			break
		}
	}
	if !exists {
		ns = append(ns, f)
	}

	return ns
}

func mapOriginResult(origin string, fr origins.FetchResult) nodes.OriginPrice {
	return nodes.OriginPrice{
		PairPrice: nodes.PairPrice{
			Pair: gofer.Pair{
				Base:  fr.Price.Pair.Base,
				Quote: fr.Price.Pair.Quote,
			},
			Price:     fr.Price.Price,
			Bid:       fr.Price.Bid,
			Ask:       fr.Price.Ask,
			Volume24h: fr.Price.Volume24h,
			Time:      fr.Price.Timestamp,
		},
		Origin: origin,
		Error:  fr.Error,
	}
}

func getMinTTL(ns []nodes.Node) time.Duration {
	minTTL := time.Duration(0)
	nodes.Walk(func(n nodes.Node) {
		if f, ok := n.(Feedable); ok {
			if minTTL == 0 || f.MinTTL() < minTTL {
				minTTL = f.MinTTL()
			}
		}
	}, ns...)

	if minTTL < time.Second {
		return time.Second
	}

	return minTTL / 2
}
