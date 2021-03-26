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
	"bytes"
	encodingJSON "encoding/json"
	"fmt"
	"time"

	"github.com/makerdao/oracle-suite/pkg/gofer"
)

type json struct {
	ndjson bool
	items  []interface{}
}

func newJSON(ndjson bool) *json {
	return &json{
		ndjson: ndjson,
	}
}

// Bytes implements the Marshaller interface.
func (j *json) Bytes() ([]byte, error) {
	buf := bytes.Buffer{}
	if j.ndjson {
		for _, item := range j.items {
			bts, err := encodingJSON.Marshal(item)
			if err != nil {
				return nil, err
			}
			buf.Write(bts)
			buf.WriteByte('\n')
		}
	} else {
		bts, err := encodingJSON.Marshal(j.items)
		if err != nil {
			return nil, err
		}
		buf.Write(bts)
		buf.WriteByte('\n')
	}
	return buf.Bytes(), nil
}

// Write implements the Marshaller interface.
func (j *json) Write(item interface{}) error {
	var i interface{}
	switch typedItem := item.(type) {
	case *gofer.Price:
		i = j.handlePrice(typedItem)
	case *gofer.Model:
		i = j.handleNode(typedItem)
	default:
		return fmt.Errorf("unsupported data type")
	}

	j.items = append(j.items, i)
	return nil
}

func (*json) handlePrice(price *gofer.Price) interface{} {
	return jsonPriceFromGoferPrice(price)
}

func (*json) handleNode(node *gofer.Model) interface{} {
	return node.Pair.String()
}

type jsonPrice struct {
	Type       string            `json:"type"`
	Base       string            `json:"base"`
	Quote      string            `json:"quote"`
	Price      float64           `json:"price"`
	Bid        float64           `json:"bid"`
	Ask        float64           `json:"ask"`
	Volume24h  float64           `json:"vol24h"`
	Timestamp  time.Time         `json:"ts"`
	Parameters map[string]string `json:"params,omitempty"`
	Prices     []jsonPrice       `json:"prices,omitempty"`
	Error      string            `json:"error,omitempty"`
}

func jsonPriceFromGoferPrice(t *gofer.Price) jsonPrice {
	var prices []jsonPrice
	for _, c := range t.Prices {
		prices = append(prices, jsonPriceFromGoferPrice(c))
	}
	return jsonPrice{
		Type:       t.Type,
		Base:       t.Pair.Base,
		Quote:      t.Pair.Quote,
		Price:      t.Price,
		Bid:        t.Bid,
		Ask:        t.Ask,
		Volume24h:  t.Volume24h,
		Timestamp:  t.Time.In(time.UTC),
		Parameters: t.Parameters,
		Prices:     prices,
		Error:      t.Error,
	}
}
