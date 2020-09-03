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

package aggregator

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/makerdao/gofer/pkg/model"
)

type Aggregator interface {
	// Add a price point to be aggregated
	Ingest(*model.PriceAggregate)
	// Calculate asset pair aggregate returning nil if pair not available
	Aggregate(*model.Pair) *model.PriceAggregate
	// GetSources returns PotentialPricePoints
	GetSources(...*model.Pair) []*model.PotentialPricePoint
}

type AggregatorParams struct {
	Name       string          `json:"name"`
	Parameters json.RawMessage `json:"parameters"`
}

type Factory func([]byte) (Aggregator, error)

var AggregatorMap = map[string]Factory{}

func FromConfig(ap AggregatorParams) (Aggregator, error) {
	af, ok := AggregatorMap[ap.Name]
	if !ok {
		return nil, fmt.Errorf("no aggregator found with name \"%s\"", ap.Name)
	}

	return af(ap.Parameters)
}

// Get price estimate from price point
func calcPrice(pp *model.PriceAggregate) float64 {
	// If ask/bid values are valid return mean of ask and bid
	if pp.Ask != 0 && pp.Bid != 0 {
		return (pp.Ask + pp.Bid) / 2
	}
	// If last auction price is valid return that
	if pp.Price != 0 {
		return pp.Price
	}
	// Otherwise return invalid price
	return 0
}

func median(xs []float64) float64 {
	count := len(xs)
	if count == 0 {
		return 0
	}

	// Sort
	sort.Slice(xs, func(i, j int) bool { return xs[i] > xs[j] })

	if count%2 == 0 {
		// Even price point count, take the mean of the two middle prices
		i := int(count / 2)
		x1 := xs[i-1]
		x2 := xs[i]
		return (x1 + x2) / 2
	}
	// Odd price point count, use the middle price
	i := int((count - 1) / 2)
	return xs[i]
}

func init() {
	AggregatorMap["setzer"] = NewSetzFromJSON
	AggregatorMap["median"] = NewMedianFromJSON
	AggregatorMap["path"] = NewPathFromJSON
}
