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

	"github.com/makerdao/gofer/pkg/gofer/graph"
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
	case graph.AggregatorTick:
		i = j.handleTick(typedItem)
	case graph.Aggregator:
		i = j.handleGraph(typedItem)
	case map[graph.Pair][]string:
		i = j.handleOrigins(typedItem)
	default:
		return fmt.Errorf("unsupported data type")
	}

	j.items = append(j.items, i)
	return nil
}

func (*json) handleTick(tick graph.AggregatorTick) interface{} {
	return jsonTickFromAggregatorTick(tick)
}

func (*json) handleGraph(graph graph.Aggregator) interface{} {
	return graph.Pair().String()
}

func (*json) handleOrigins(origins map[graph.Pair][]string) interface{} {
	r := make(map[string][]string)
	for p, o := range origins {
		r[p.String()] = o
	}
	return r
}

type jsonTick struct {
	Type       string            `json:"type,omitempty"`
	Parameters map[string]string `json:"params,omitempty"`
	Origin     string            `json:"origin,omitempty"`
	Base       string            `json:"base,omitempty"`
	Quote      string            `json:"quote,omitempty"`
	Price      float64           `json:"price,omitempty"`
	Bid        float64           `json:"bid,omitempty"`
	Ask        float64           `json:"ask,omitempty"`
	Volume24h  float64           `json:"vol24h,omitempty"`
	Timestamp  time.Time         `json:"ts,omitempty"`
	Ticks      []jsonTick        `json:"ticks,omitempty"`
	Error      string            `json:"error,omitempty"`
}

func jsonTickFromOriginTick(t graph.OriginTick) jsonTick {
	var errStr string
	if t.Error != nil {
		errStr = t.Error.Error()
	}

	return jsonTick{
		Type:       "origin",
		Parameters: nil,
		Origin:     t.Origin,
		Base:       t.Pair.Base,
		Quote:      t.Pair.Quote,
		Price:      t.Price,
		Bid:        t.Bid,
		Ask:        t.Ask,
		Volume24h:  t.Volume24h,
		Timestamp:  t.Timestamp.In(time.UTC),
		Error:      errStr,
	}
}

func jsonTickFromAggregatorTick(t graph.AggregatorTick) jsonTick {
	var errStr string
	if t.Error != nil {
		errStr = t.Error.Error()
	}

	var ticks []jsonTick
	for _, v := range t.OriginTicks {
		ticks = append(ticks, jsonTickFromOriginTick(v))
	}
	for _, v := range t.AggregatorTicks {
		ticks = append(ticks, jsonTickFromAggregatorTick(v))
	}

	return jsonTick{
		Type:       "aggregate",
		Parameters: t.Parameters,
		Base:       t.Pair.Base,
		Quote:      t.Pair.Quote,
		Price:      t.Price,
		Bid:        t.Bid,
		Ask:        t.Ask,
		Volume24h:  t.Volume24h,
		Timestamp:  t.Timestamp.In(time.UTC),
		Ticks:      ticks,
		Error:      errStr,
	}
}
