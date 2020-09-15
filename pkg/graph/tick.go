package graph

import (
	"fmt"
	"strings"
	"time"
)

type Pair struct {
	Quote string
	Base  string
}

func NewPair(s string) (Pair, error) {
	ss := strings.Split(s, "/")
	if len(ss) != 2 {
		return Pair{}, fmt.Errorf("couldn't parse pair \"%s\"", s)
	}

	return Pair{Base: strings.ToUpper(ss[0]), Quote: strings.ToUpper(ss[1])}, nil
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
