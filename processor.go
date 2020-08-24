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
	"log"

	"github.com/makerdao/gofer/aggregator"
	"github.com/makerdao/gofer/exchange"
	"github.com/makerdao/gofer/model"
)

type Processor struct {
	exchangeSet *exchange.ExchangesSet
}

// NewProcessor instantiate new `Processor` instance with custom `query.WorkerPool`
func NewProcessor(set *exchange.ExchangesSet) *Processor {
	return &Processor{
		exchangeSet: set,
	}
}

// ProcessOne processes `PotentialPricePoint` and fetches new price for it
func (p *Processor) ProcessOne(pp *model.PotentialPricePoint) (*model.PriceAggregate, error) {
	if err := model.ValidatePotentialPricePoint(pp); err != nil {
		return nil, err
	}
	point, err := p.exchangeSet.Call(pp)
	if err != nil {
		return nil, err
	}
	return &model.PriceAggregate{
		PriceModelName: fmt.Sprintf("exchange[%s]", pp.Exchange.Name),
		PricePoint:     point,
	}, nil
}

// Process takes `PotentialPricePoint` as an input fetches all required info using `query`
// system, passes everything to given `aggregator` and returns it.
// Technically you don't even need to get passed `aggregator` back, because you can use pointer to passed one.
// and here it returns just for clearer API.
func (p *Processor) Process(pairs []*model.Pair, agg aggregator.Aggregator) (aggregator.Aggregator, error) {
	if agg == nil {
		return nil, fmt.Errorf("no working agregator passed to processor")
	}
	for _, pp := range agg.GetSources(pairs) {
		res, err := p.ProcessOne(pp)
		if err != nil {
			// TODO: log exchange errors here so failures are traceable but does't fail
			// everything because of a single bad exchange reply
			log.Println(err)
			continue
		}
		agg.Ingest(res)
	}
	return agg, nil
}
