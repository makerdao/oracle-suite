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

	"github.com/makerdao/gofer/model"
)

type IdentityPather struct {
	PricePathMap model.PricePathMap
}

func (ip *IdentityPather) Pairs() []*model.Pair {
	var pairs []*model.Pair
	for p := range ip.PricePathMap {
		pairs = append(pairs, &p)
	}
	return pairs
}

func (ip *IdentityPather) Path(target *model.Pair) []*model.PricePath {
	return ip.PricePathMap[*target]
}

// Path is an aggregator that resolves price paths for indirect pairs and takes
// the median of all paths for each pair
type Path struct {
	pather           Pather
	directAggregator Aggregator
	sources          []*model.PotentialPricePoint
}

func paths(pather Pather, pairs []*model.Pair) []*model.PricePath {
	pairs_ := make(map[model.Pair]bool)
	for _, p := range pairs {
		pairs_[*p] = true
	}
	var ppaths []*model.PricePath
	for pair := range pairs_ {
		ppaths_ := pather.Path(&pair)
		if ppaths_ != nil {
			ppaths = append(ppaths, ppaths_...)
		}
	}
	return ppaths
}

// NewPath returns a new instance of `Path` that uses the given price paths to
// aggregate indirect pairs and an aggregator to merge direct pairs.
func NewPath(pather Pather, sources []*model.PotentialPricePoint, directAggregator Aggregator) *Path {
	return &Path{
		pather:           pather, // *model.NewPricePathMap(ppaths),
		directAggregator: directAggregator,
		sources:          sources,
	}
}

func NewPathWithPathMap(ppaths []*model.PricePath, sources []*model.PotentialPricePoint, directAggregator Aggregator) *Path {
	pather := &IdentityPather{PricePathMap: model.NewPricePathMap(ppaths)}
	return NewPath(
		pather,
		sources,
		directAggregator,
	)
}

type PricePath []Pair

func ToModelPricePaths(pps []PricePath) []*model.PricePath {
	var ppaths []*model.PricePath
	for _, pp := range pps {
		var ppath model.PricePath
		for _, p := range pp {
			ppath = append(ppath, &p.Pair)
		}
		ppaths = append(ppaths, &ppath)
	}
	return ppaths
}

type PathParams struct {
	PricePaths       []PricePath      `json:"paths"`
	DirectAggregator AggregatorParams `json:"aggregator"`
	Sources          []Source         `json:"sources"`
}

func NewPathFromJSON(raw []byte) (Aggregator, error) {
	var params PathParams
	err := json.Unmarshal(raw, &params)
	if err != nil {
		return nil, fmt.Errorf("couldn't parse path aggregator parameters: %w", err)
	}

	subAgg, err := FromConfig(params.DirectAggregator)
	if err != nil {
		return nil, err
	}

	ppps := ToPotentialPricePoints(params.Sources)

	return NewPathWithPathMap(ToModelPricePaths(params.PricePaths), ppps, subAgg), nil
}

// Calculate the final model.trade price of an ordered list of prices
func trade(pas []*model.PriceAggregate) *model.PriceAggregate {
	var pair *model.Pair
	var price float64

	for _, pa := range pas {
		if price == 0 {
			price = pa.Price
			pair = pa.Pair.Clone()
		} else if pair.Base == pa.Pair.Base {
			price = pa.Price / price
			pair.Base = pair.Quote
			pair.Quote = pa.Pair.Quote
		} else {
			price *= pa.Price
			pair.Quote = pa.Pair.Quote
		}
	}

	return model.NewPriceAggregate(
		"trade",
		&model.PricePoint{
			Pair:  pair,
			Price: price,
		},
		pas...,
	)
}

func (r *Path) resolve(ppath model.PricePath) *model.PriceAggregate {
	var pas []*model.PriceAggregate
	for _, pair := range ppath {
		pa := r.directAggregator.Aggregate(pair)
		if pa == nil {
			return nil
		}

		pas = append(pas, pa)
	}
	return trade(pas)
}

func (r *Path) Ingest(pa *model.PriceAggregate) {
	r.directAggregator.Ingest(pa)
}

func (r *Path) Aggregate(pair *model.Pair) *model.PriceAggregate {
	if pair == nil {
		return nil
	}

	ppaths := r.pather.Path(pair)
	if ppaths == nil {
		return nil
	}

	var pas []*model.PriceAggregate
	var prices []float64
	for _, path := range ppaths {
		if pa := r.resolve(*path); pa != nil {
			pas = append(pas, pa)
			prices = append(prices, pa.Price)
		}
	}

	return model.NewPriceAggregate(
		"path",
		&model.PricePoint{
			Pair:  pair.Clone(),
			Price: median(prices),
		},
		pas...,
	)
}

func (r *Path) GetSources(pairs []*model.Pair) []*model.PotentialPricePoint {
	ppaths := paths(r.pather, pairs)
	_, ppps := FilterPotentialPricePoints(ppaths, r.sources)
	return ppps
}
