package graph

import (
	"fmt"
	"time"

	"github.com/makerdao/gofer/pkg/exchange"
	"github.com/makerdao/gofer/pkg/model"
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
			if ingestableNode.Tick().Timestamp.Before(t.Add(time.Second * time.Duration(-1 * i.ttl))) {
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

// TODO: Temp function, to remove
func (i *Ingestor) fetch(ep ExchangePair) ExchangeTick {
	p, _ := model.NewPairFromString(ep.Pair.Base + "/" + ep.Pair.Quote)
	ppp := &model.PotentialPricePoint{
		Pair: p,
		Exchange: &model.Exchange{
			Name:   ep.Exchange,
			Config: nil,
		},
	}

	cr := i.set.Call([]*model.PotentialPricePoint{ppp})

	if len(cr) != 1 {
		return ExchangeTick{
			Tick: Tick{
				Pair: ep.Pair,
			},
			Exchange: ep.Exchange,
			Error:    fmt.Errorf("unable to fetch tick for %s", ep.Pair),
		}
	}

	if cr[0].Error != nil {
		return ExchangeTick{
			Tick: Tick{
				Pair: ep.Pair,
			},
			Exchange: ep.Exchange,
			Error:    cr[0].Error,
		}
	}

	return ExchangeTick{
		Tick: Tick{
			Pair:      ep.Pair,
			Price:     cr[0].PricePoint.Price,
			Bid:       cr[0].PricePoint.Bid,
			Ask:       cr[0].PricePoint.Ask,
			Volume24h: cr[0].PricePoint.Volume,
			Timestamp: time.Unix(cr[0].PricePoint.Timestamp, 0),
		},
		Exchange: ep.Exchange,
		Error:    cr[0].Error,
	}
}
