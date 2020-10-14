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
	"fmt"
	"strings"
	"time"
)

type Pair struct {
	Base  string
	Quote string
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
	return p.Base == c.Base && p.Quote == c.Quote
}

func (p Pair) String() string {
	return fmt.Sprintf("%s/%s", p.Base, p.Quote)
}

type OriginPair struct {
	Origin string
	Pair   Pair
}

func (o OriginPair) String() string {
	return fmt.Sprintf("%s %s", o.Pair.String(), o.Origin)
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
	Origin string // Origin is a name of Tick source.
	Error  error  // Error is optional error which may occur during fetching Tick.
}

// AggregatorTick represent Tick which was calculated using other ticks.
type AggregatorTick struct {
	Tick
	OriginTicks     []OriginTick      // OriginTicks is a list of all OriginTicks used to calculate Tick.
	AggregatorTicks []AggregatorTick  // AggregatorTicks is a list of all OriginTicks used to calculate Tick.
	Parameters      map[string]string // Parameters is a custom list of optional parameters returned by an aggregator.
	Error           error             // Error is optional error which may occur during calculating Tick.
}

func Pairs(l PriceModels, args ...string) ([]Pair, error) {
	var pairs []Pair
	if len(args) > 0 {
		for _, pair := range args {
			p, err := NewPair(pair)
			if err != nil {
				return nil, err
			}
			pairs = append(pairs, p)
		}
	} else {
		pairs = l.Pairs()
	}
	return pairs, nil
}
