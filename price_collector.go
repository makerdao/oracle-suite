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
)

// PriceCollector will collect prices for you
type PriceCollector struct {
	wp query.WorkerPool
}

// NewPriceCollector create new ready to work `PriceCollector`
func NewPriceCollector() *PriceCollector {
	workerPool := query.NewHTTPWorkerPool(10)
	workerPool.Start()

	return &PriceCollector{
		wp: workerPool,
	}
}

// CollectPricePoint makes request to exchange and fetching a price point
func (pc *PriceCollector) CollectPricePoint(pp *model.PotentialPricePoint) (*model.PricePoint, error) {
	if pc.wp == nil || !pc.wp.Ready() {
		return nil, fmt.Errorf("wrong worker pool defined for PriceCollector")
	}

	// Making a call
	return exchange.Call(pc.wp, pp)
}
