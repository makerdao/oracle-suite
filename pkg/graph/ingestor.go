package graph

import (
	"fmt"
	"time"

	"github.com/makerdao/gofer/pkg/exchange"
)

type Ingestable interface {
	ExchangePair() ExchangePair
	SetTick(tick ExchangeTick)
	Tick() ExchangeTick
}

// Ingestor sets data to the Ingestable nodes.
type Ingestor struct {
	set *exchange.Set
	ttl int
}

func NewIngestor(set *exchange.Set, ttl int) *Ingestor {
	return &Ingestor{set: set, ttl: ttl}
}

func (i *Ingestor) Ingest(node Node) {
	t := time.Now()

	AsyncWalk(node, func(node Node) {
		if ingestableNode, ok := node.(Ingestable); ok {
			if ingestableNode.Tick().Timestamp.Before(t.Add(time.Second * time.Duration(-1*i.ttl))) {
				ingestableNode.SetTick(i.fetch(ingestableNode.ExchangePair()))
			}
		}
	})
}

func (i *Ingestor) Clear(node Node) {
	Walk(node, func(node Node) {
		if ingestableNode, ok := node.(Ingestable); ok {
			ingestableNode.SetTick(ExchangeTick{})
		}
	})
}

func (i *Ingestor) fetch(ep ExchangePair) ExchangeTick {
	// TODO: update for batch requests
	crs := i.set.Call(map[string][]exchange.Pair{
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
