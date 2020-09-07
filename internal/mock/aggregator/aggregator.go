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
	"github.com/makerdao/gofer/pkg/model"
)

type Aggregator struct {
	Returns map[model.Pair]*model.PriceAggregate
	Sources map[model.Pair][]*model.PricePoint
}

func (mr *Aggregator) Ingest(pa *model.PriceAggregate) {
}

func (mr *Aggregator) Aggregate(pair *model.Pair) *model.PriceAggregate {
	if pair == nil {
		return nil
	}
	return mr.Returns[*pair]
}

func (mr *Aggregator) GetSources(pairs ...*model.Pair) []*model.PricePoint {
	if pairs == nil {
		for p := range mr.Sources {
			pairs = append(pairs, p.Clone())
		}
	}
	pps := make(map[string]*model.PricePoint)
	for _, p := range pairs {
		for _, pp := range mr.Sources[*p] {
			pps[pp.String()] = pp
		}
	}
	var sources []*model.PricePoint
	for _, pp := range pps {
		sources = append(sources, pp)
	}
	return sources
}
