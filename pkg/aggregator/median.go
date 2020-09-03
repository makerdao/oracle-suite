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

	"github.com/makerdao/gofer/pkg/model"
)

type traceAggregate struct {
	*model.PriceAggregate
	prices          map[string]*model.PriceAggregate
	newestTimestamp int64
	reduced         bool
}

type Median struct {
	timeWindow int64
	aggregates map[model.Pair]*traceAggregate
	sources    []*model.PotentialPricePoint
}

// NewMedian returns a new instance of Medain with a time window in milliseconds
func NewMedian(sources []*model.PotentialPricePoint, timeWindow int64) *Median {
	return &Median{
		timeWindow: timeWindow,
		aggregates: make(map[model.Pair]*traceAggregate),
		sources:    sources,
	}
}

type Source struct {
	Base       string            `json:"base"`
	Quote      string            `json:"quote"`
	Exchange   string            `json:"exchange"`
	Parameters map[string]string `json:"parameters"`
}

func ToPotentialPricePoints(sources []Source) []*model.PotentialPricePoint {
	var ppps []*model.PotentialPricePoint
	for _, s := range sources {
		ppps = append(ppps, &model.PotentialPricePoint{
			Pair:     model.NewPair(s.Base, s.Quote),
			Exchange: &model.Exchange{Name: s.Exchange, Config: s.Parameters},
		})
	}
	return ppps
}

type MedianParams struct {
	TimeWindow int64    `json:"timewindow"`
	Sources    []Source `json:"sources"`
}

// NewMedian returns a new instance of Medain with a time window in milliseconds
func NewMedianFromJSON(raw []byte) (Aggregator, error) {
	var params MedianParams
	err := json.Unmarshal(raw, &params)
	if err != nil {
		return nil, fmt.Errorf("couldn't parse median aggregator parameters: %w", err)
	}

	return NewMedian(ToPotentialPricePoints(params.Sources), params.TimeWindow), nil
}

func newTraceAggregate(pair *model.Pair) *traceAggregate {
	return &traceAggregate{
		PriceAggregate:  model.NewPriceAggregate("median", &model.PricePoint{Pair: pair}),
		newestTimestamp: 0,
		prices:          make(map[string]*model.PriceAggregate),
		reduced:         false,
	}
}

// Add a price point to median reducer state
func (r *Median) Ingest(pa *model.PriceAggregate) {
	// Ignore if input is nil
	if pa == nil {
		return
	}

	// Ignore price point if no valid price
	price := calcPrice(pa)
	if price == 0 {
		return
	}

	pair := *pa.Pair
	// Create new trace aggregate if one doesn't already exist for asset pair
	if _, ok := r.aggregates[pair]; !ok {
		r.aggregates[pair] = newTraceAggregate(pa.Pair.Clone())
	}
	trace := r.aggregates[pair]

	if len(trace.prices) == 0 || pa.Timestamp > trace.newestTimestamp {
		trace.newestTimestamp = pa.Timestamp
	}

	timeWindow := trace.newestTimestamp - r.timeWindow
	// New price is outside time window, do nothing
	if pa.Timestamp <= timeWindow {
		return
	}

	existingPrice := trace.prices[pa.Exchange.Name]
	// Price with same exchange as new price already exists
	if existingPrice == nil || pa.Timestamp > existingPrice.Timestamp {
		// Update existing price if new price is newer
		trace.prices[pa.Exchange.Name] = pa
		// Set state to dirty
		trace.reduced = false
	}
}

// Sort prices in state and return median
func (r *Median) Aggregate(pair *model.Pair) *model.PriceAggregate {
	if pair == nil {
		return nil
	}

	trace := r.aggregates[*pair]
	if trace == nil {
		return nil
	}

	if trace.reduced || len(trace.prices) == 0 {
		return trace.Clone()
	}

	timeWindow := trace.newestTimestamp - r.timeWindow
	var pas []*model.PriceAggregate
	for _, p := range trace.prices {
		// Only add prices inside time window
		if p.Timestamp > timeWindow {
			pas = append(pas, p)
		} else {
			delete(trace.prices, p.Exchange.Name)
		}
	}

	prices := make([]float64, len(pas))
	for i, pa := range pas {
		prices[i] = calcPrice(pa)
	}
	trace.Price = median(prices)
	trace.Prices = pas
	trace.reduced = true
	return trace.Clone()
}

func (r *Median) GetSources(pairs ...*model.Pair) []*model.PotentialPricePoint {
	pairMap := make(map[model.Pair]bool)
	for _, pair := range pairs {
		pairMap[*pair] = true
	}

	var ppps []*model.PotentialPricePoint
	for _, ppp := range r.sources {
		if _, ok := pairMap[*ppp.Pair]; ok {
			ppps = append(ppps, ppp)
		}
	}
	return ppps
}
