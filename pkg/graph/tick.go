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

type OriginPair struct {
	Origin string
	Pair   Pair
}

type Tick struct {
	Pair      Pair
	Price     float64
	Bid       float64
	Ask       float64
	Volume24h float64
	Timestamp time.Time
}

// OriginTick represent Tick which was sourced directly from an origin.
type OriginTick struct {
	Tick
	Origin string
	Error  error
}

// IndirectTick represent Tick which was calculated using other ticks.
type IndirectTick struct {
	Tick
	OriginTicks   []OriginTick
	IndirectTicks []IndirectTick
	Error         error
}
