package graph

import (
	"time"

	"github.com/makerdao/gofer/pkg/exchange"
	"github.com/makerdao/gofer/pkg/model"
)

type Ingestable interface {
	ExchangePair() ExchangePair
	SetTick(tick ExchangeTick)
}

// Ingestor sets data to the Ingestable nodes.
type Ingestor struct {
	set *exchange.Set
}

func NewIngestor(set *exchange.Set) *Ingestor {
	return &Ingestor{set: set}
}

func (i *Ingestor) Ingest(node Node) {
	AsyncWalk(node, func(node Node) {
		if ingestableNode, ok := node.(Ingestable); ok {
			ingestableNode.SetTick(i.fetch(ingestableNode.ExchangePair()))
		}
	})
}

func (i *Ingestor) Clear(node Node) {
	AsyncWalk(node, func(node Node) {
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
