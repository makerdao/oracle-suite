package graph

import (
	"fmt"
	"time"

	"github.com/makerdao/gofer/pkg/origins"
)

type Feedable interface {
	OriginPair() OriginPair
	Feed(tick OriginTick)
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

func (i *Feeder) Feed(node Node) {
	t := time.Now()

	AsyncWalk(node, func(node Node) {
		if feedable, ok := node.(Feedable); ok {
			if feedable.Tick().Timestamp.Before(t.Add(time.Second * time.Duration(-1*i.ttl))) {
				feedable.Feed(i.fetch(feedable.OriginPair()))
			}
		}
	})
}

func (i *Feeder) Clear(node Node) {
	Walk(node, func(node Node) {
		if feedable, ok := node.(Feedable); ok {
			feedable.Feed(OriginTick{})
		}
	})
}

func (i *Feeder) fetch(ep OriginPair) OriginTick {
	// TODO: update for batch requests
	crs := i.set.Fetch(map[string][]origins.Pair{
		ep.Origin: {{
			Quote: ep.Pair.Quote,
			Base:  ep.Pair.Base,
		}},
	})

	if len(crs[ep.Origin]) != 1 {
		return OriginTick{
			Tick: Tick{
				Pair: ep.Pair,
			},
			Origin: ep.Origin,
			Error:  fmt.Errorf("unable to fetch tick for %s", ep.Pair),
		}
	}

	cr := crs[ep.Origin][0]

	if cr.Error != nil {
		return OriginTick{
			Tick: Tick{
				Pair: ep.Pair,
			},
			Origin: ep.Origin,
			Error:  cr.Error,
		}
	}

	return OriginTick{
		Tick: Tick{
			Pair:      ep.Pair,
			Price:     cr.Tick.Price,
			Bid:       cr.Tick.Bid,
			Ask:       cr.Tick.Ask,
			Volume24h: cr.Tick.Volume24h,
			Timestamp: cr.Tick.Timestamp,
		},
		Origin: ep.Origin,
		Error:  cr.Error,
	}
}
