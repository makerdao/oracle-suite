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
	"log"

	"github.com/makerdao/gofer/pkg/model"
)

type Setz struct {
	pairMap     PriceModelMap
	exchangeMap map[string]*model.Exchange
	cache       PriceCache
}

func NewSetz(exchanges []*model.Exchange, pairMap PriceModelMap) *Setz {
	exchangeMap := make(map[string]*model.Exchange)
	for _, e := range exchanges {
		exchangeMap[e.Name] = e
	}
	return &Setz{
		pairMap:     pairMap,
		exchangeMap: exchangeMap,
		cache:       make(PriceCache),
	}
}

type SetzParams struct {
	// TODO: change this into map[string]model.Exchange to allow aliasing of exchange configs
	Exchanges   map[string]map[string]string `json:"origins"`
	PriceModels PriceModelMap                `json:"pricemodels"`
}

func NewSetzFromJSON(raw []byte) (Aggregator, error) {
	var params SetzParams
	err := json.Unmarshal(raw, &params)
	if err != nil {
		return nil, fmt.Errorf("couldn't parse setz aggregator parameters: %w", err)
	}

	var exchanges []*model.Exchange
	for name, eparams := range params.Exchanges {
		exchanges = append(exchanges, &model.Exchange{Name: name, Config: eparams})
	}

	return NewSetz(exchanges, params.PriceModels), nil
}

func (a *Setz) Ingest(pa *model.PriceAggregate) {
	alias := cacheID{pair: Pair{*pa.Pair}, exchangeName: pa.Exchange.Name}
	// Copy PriceAggregate into cache
	a.cache[alias] = pa.Clone()
}

func (a *Setz) Aggregate(pair *model.Pair) *model.PriceAggregate {
	if pair == nil {
		return nil
	}

	pa, err := a.pairMap.ResolveRef(a.cache, PriceRef{Origin: ".", Pair: Pair{*pair}})
	if err != nil {
		// TODO: refactor Aggregator to return error not just nil
		log.Println(err)
		return nil
	}

	return pa
}

func (a *Setz) GetSources(pairs ...*model.Pair) []*model.PotentialPricePoint {
	// If given list of pairs is empty use all keys in PriceModelMap as pair list
	if pairs == nil {
		for p := range a.pairMap {
			pairs = append(pairs, p.Clone())
		}
	}

	var refs []PriceRef
	for _, p := range pairs {
		refs = append(refs, PriceRef{Origin: ".", Pair: Pair{*p}})
	}
	ppps, err := a.pairMap.GetRefSources(a.exchangeMap, refs...)
	if err != nil {
		// TODO: refactor Aggregator to return error not just nil
		log.Println(err)
		return nil
	}

	return ppps
}
