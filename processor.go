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
	"fmt"
	"makerdao/gofer/exchange"
	"makerdao/gofer/model"
	"makerdao/gofer/query"
	"makerdao/gofer/reducer"
)

type Processor struct {
	wp query.WorkerPool
}

// NewProcessor instantiate new `Processor` instance with custom `query.WorkerPool`
func NewProcessor(wp query.WorkerPool) *Processor {
	return &Processor{
		wp: wp,
	}
}

func NewProcessorWithHTTPWorkerPool() *Processor {
	p := &Processor{
		wp: query.NewHTTPWorkerPool(5),
	}
	p.wp.Start()
	return p
}

// Process takes `PotentialPricePoint` as an input fetches all required info using `query`
// system, passes everything to `reducer` and returns result.
func (p *Processor) Process(pp *model.PotentialPricePoint) (*model.PriceAggregate, error) {
	if p.wp == nil || !p.wp.Ready() {
		return nil, fmt.Errorf("worker pool is not ready for querying prices")
	}
	if err := model.ValidatePotentialPricePoint(pp); err != nil {
		return nil, err
	}
	point, err := exchange.Call(p.wp, pp)
	if err != nil {
		return nil, err
	}
	// TODO: wrong usage, need to define timeWindow & reducer type
	medianReducer := reducer.NewMedianReducer(pp.Pair, 300)
	medianReducer.Ingest(point)

	return medianReducer.Reduce(), nil
}