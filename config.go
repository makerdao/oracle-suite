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

package gofer

import (
	"makerdao/gofer/aggregator"
	"makerdao/gofer/model"
	"makerdao/gofer/pather"
)

type AggregateProcessor interface {
	Process(ppps []*model.PotentialPricePoint, agg aggregator.Aggregator) (aggregator.Aggregator, error)
}

type Config struct {
	Sources       []*model.PotentialPricePoint
	NewAggregator func([]*model.PricePath) aggregator.Aggregator
	Pather        pather.Pather
	Processor     AggregateProcessor
}

func NewConfig(sources []*model.PotentialPricePoint, newAggregator func([]*model.PricePath) aggregator.Aggregator, pather pather.Pather, processor AggregateProcessor) *Config {
	return &Config{
		Sources:       sources,
		NewAggregator: newAggregator,
		Pather:        pather,
		Processor:     processor,
	}
}

func NewConfigWithDefaults(sources []*model.PotentialPricePoint) *Config {
	return NewConfig(
		sources,
		func(ppaths []*model.PricePath) aggregator.Aggregator {
			return aggregator.NewPath(ppaths, aggregator.NewMedian(1000))
		},
		pather.NewSetzer(),
		NewProcessorWithHTTPWorkerPool(),
	)
}
