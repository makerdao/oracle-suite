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

package marshal

import (
	encodingJson "encoding/json"
	"fmt"

	"github.com/makerdao/gofer/pkg/aggregator"
	"github.com/makerdao/gofer/pkg/model"
)

type json struct {
	bufferedMarshaller *bufferedMarshaller
}

func newJSON(ndjson bool) *json {
	return &json{newBufferedMarshaller(ndjson, func(item interface{}, ierr error) ([]marshalledItem, error) {
		if i, ok := item.([]marshalledItem); ok {
			b, err := encodingJson.Marshal(i)
			b = append(b, '\n')
			return []marshalledItem{b}, err
		}

		var err error
		var ret []marshalledItem

		switch i := item.(type) {
		case *model.PriceAggregate:
			err = jsonHandlePriceAggregate(&ret, i, ierr)
		case *model.Exchange:
			err = jsonHandleExchange(&ret, i)
		case aggregator.PriceModelMap:
			err = jsonHandlePriceModelMap(&ret, i)
		default:
			return nil, fmt.Errorf("unsupported data type")
		}

		return ret, err
	})}
}

// Read implements the Marshaller interface.
func (j *json) Read(p []byte) (int, error) {
	return j.bufferedMarshaller.Read(p)
}

// Write implements the Marshaller interface.
func (j *json) Write(item interface{}, err error) error {
	return j.bufferedMarshaller.Write(item, err)
}

// Close implements the Marshaller interface.
func (j *json) Close() error {
	return j.bufferedMarshaller.Close()
}

func jsonHandlePriceAggregate(ret *[]marshalledItem, aggregate *model.PriceAggregate, err error) error {
	b, err := encodingJson.Marshal(newPriceAggregate(aggregate, err))
	if err != nil {
		return err
	}

	b = append(b, '\n')
	*ret = append(*ret, b)
	return nil
}

func jsonHandleExchange(ret *[]marshalledItem, exchange *model.Exchange) error {
	b, err := encodingJson.Marshal(exchange.Name)
	if err != nil {
		return err
	}

	b = append(b, '\n')
	*ret = append(*ret, b)
	return nil
}

func jsonHandlePriceModelMap(ret *[]marshalledItem, priceModelMap aggregator.PriceModelMap) error {
	for pr, pm := range priceModelMap {
		b, err := encodingJson.Marshal(struct {
			Pair  aggregator.Pair
			Model aggregator.PriceModel
		}{pr, pm})
		if err != nil {
			return err
		}

		b = append(b, '\n')
		*ret = append(*ret, b)
	}
	return nil
}
