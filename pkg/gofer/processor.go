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

	"github.com/makerdao/gofer/pkg/aggregator"
	"github.com/makerdao/gofer/pkg/exchange"
	"github.com/makerdao/gofer/pkg/model"
)

type Processor struct {
	exchangeSet *exchange.Set
}

// NewProcessor instantiate new `Processor` instance with custom `query.WorkerPool`
func NewProcessor(set *exchange.Set) *Processor {
	return &Processor{
		exchangeSet: set,
	}
}

// Process takes `PotentialPricePoint` as an input fetches all required info using `query`
// system, passes everything to given `aggregator` and returns it.
// Technically you don't even need to get passed `aggregator` back, because you can use pointer to passed one.
// and here it returns just for clearer API.
func (p *Processor) Process(pairs []*model.Pair, agg aggregator.Aggregator) (aggregator.Aggregator, error) {
	if agg == nil {
		return nil, fmt.Errorf("no working agregator passed to processor")
	}

	for _, cr := range p.exchangeSet.Call(agg.GetSources(pairs)) {
		if cr.Error != nil {
			// TODO: log exchange errors here so failures are traceable but does't fail
			// everything because of a single bad exchange reply
			log.Println(cr.Error)
			continue
		}

		pa := &model.PriceAggregate{
			PriceModelName: fmt.Sprintf("exchange[%s]", cr.PricePoint.Exchange.Name),
			PricePoint:     cr.PricePoint,
		}

		agg.Ingest(pa)
	}

	return agg, nil
}
