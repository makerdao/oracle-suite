package graph

import (
	"fmt"
	"time"

	"github.com/makerdao/gofer/pkg/origins"
)

type Feedable interface {
	ExchangePair() ExchangePair
	Feed(tick ExchangeTick)
	Tick() ExchangeTick
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
				feedable.Feed(i.fetch(feedable.ExchangePair()))
			}
		}
	})
}

func (i *Feeder) Clear(node Node) {
	Walk(node, func(node Node) {
		if ingestableNode, ok := node.(Feedable); ok {
			ingestableNode.Feed(ExchangeTick{})
		}
	})
}

func (i *Feeder) fetch(ep ExchangePair) ExchangeTick {
	// TODO: update for batch requests
	crs := i.set.Call(map[string][]origins.Pair{
		ep.Exchange: {{
			Quote: ep.Pair.Quote,
			Base:  ep.Pair.Base,
		}},
	})

	if len(crs[ep.Exchange]) != 1 {
		return ExchangeTick{
			Tick: Tick{
				Pair: ep.Pair,
			},
			Exchange: ep.Exchange,
			Error:    fmt.Errorf("unable to fetch tick for %s", ep.Pair),
		}
	}

	cr := crs[ep.Exchange][0]

	if cr.Error != nil {
		return ExchangeTick{
			Tick: Tick{
				Pair: ep.Pair,
			},
			Exchange: ep.Exchange,
			Error:    cr.Error,
		}
	}

	return ExchangeTick{
		Tick: Tick{
			Pair:      ep.Pair,
			Price:     cr.Tick.Price,
			Bid:       cr.Tick.Bid,
			Ask:       cr.Tick.Ask,
			Volume24h: cr.Tick.Volume24h,
			Timestamp: cr.Tick.Timestamp,
		},
		Exchange: ep.Exchange,
		Error:    cr.Error,
	}
}
