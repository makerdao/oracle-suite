package graph

import (
	"fmt"
	"time"
)

type Aggregator interface {
	Node
	Tick() IndirectTick
}

type Exchange interface {
	Node
	Tick() ExchangeTick
}

type Pair struct {
	Quote string
	Base  string
}

func (p Pair) Empty() bool {
	return p.Base == "" && p.Quote == ""
}

func (p Pair) Equal(c Pair) bool {
	return p.Quote == c.Quote && p.Base == c.Base
}

func (p Pair) String() string {
	return fmt.Sprintf("%s/%s", p.Base, p.Quote)
}

type ExchangePair struct {
	Exchange string
	Pair     Pair
}

type Tick struct {
	Pair      Pair
	Price     float64
	Bid       float64
	Ask       float64
	Volume24h float64
	Timestamp time.Time
}

// ExchangeTick represent Tick which was sourced directly from an exchange.
type ExchangeTick struct {
	Tick
	Exchange string
	Error    error
}

// IndirectTick represent Tick which was calculated using other ticks.
type IndirectTick struct {
	Tick
	ExchangeTicks []ExchangeTick
	IndirectTick  []IndirectTick
	Error         error
}
