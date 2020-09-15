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

// AggregatorTick represent Tick which was calculated using other ticks.
type AggregatorTick struct {
	Tick
	OriginTicks     []OriginTick
	AggregatorTicks []AggregatorTick
	Method          string
	Error           error
}
