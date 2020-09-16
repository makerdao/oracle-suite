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

	"github.com/makerdao/gofer/pkg/origins"
)

type Feedable interface {
	OriginPair() OriginPair
	Ingest(tick OriginTick)
	Tick() OriginTick
}

// Feeder sets data to the Feedable nodes.
type Feeder struct {
	set *origins.Set
	ttl int
}

func NewFeeder(set *origins.Set, ttl int) *Feeder {
	return &Feeder{set: set, ttl: ttl}
}

func (i *Feeder) Feed(nodes ...Node) {
	i.fetchTicks(i.findFeedableNodes(nodes))
}

func (i *Feeder) Clear(nodes ...Node) {
	Walk(func(node Node) {
		if feedable, ok := node.(Feedable); ok {
			feedable.Ingest(OriginTick{})
		}
	}, nodes...)
}

func (i *Feeder) findFeedableNodes(nodes []Node) []Feedable {
	t := time.Now().Add(time.Second * time.Duration(-1*i.ttl))

	var feedables []Feedable
	Walk(func(node Node) {
		if feedable, ok := node.(Feedable); ok {
			if feedable.Tick().Timestamp.Before(t) {
				feedables = append(feedables, feedable)
			}
		}
	}, nodes...)

	return feedables
}

func (i *Feeder) fetchTicks(nodes []Feedable) {
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

			for _, node := range nodesMap[originPair] {
				node.Ingest(mapOriginResult(origin, fr))
			}
		}
	}
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
