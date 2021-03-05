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
	"time"

	"github.com/makerdao/gofer/pkg/gofer"
)

type OriginPair struct {
	Origin string
	Pair   gofer.Pair
}

func (o OriginPair) String() string {
	return fmt.Sprintf("%s %s", o.Pair.String(), o.Origin)
}

type Tick struct {
	Pair      gofer.Pair
	Price     float64
	Bid       float64
	Ask       float64
	Volume24h float64
	Time      time.Time
}

// OriginTick represent Tick which was sourced directly from an origin.
type OriginTick struct {
	Tick
	// Origin is a name of Tick source.
	Origin string
	// Errors is a list of optional error messages which may occur during
	// calculating Tick. If this list is not empty, then the tick value
	// is not reliable.
	Error error
}

// AggregatorTick represent Tick which was calculated using other ticks.
type AggregatorTick struct {
	Tick
	// OriginTicks is a list of all OriginTicks used to calculate Tick.
	OriginTicks []OriginTick
	// AggregatorTicks is a list of all OriginTicks used to calculate Tick.
	AggregatorTicks []AggregatorTick
	// Parameters is a custom list of optional parameters returned by an aggregator.
	Parameters map[string]string
	// Errors is a list of optional error messages which may occur during
	// fetching Tick. If this list is not empty, then the tick value
	// is not reliable.
	Error error
}
