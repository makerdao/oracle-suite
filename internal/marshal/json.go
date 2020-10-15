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
	"time"

	"github.com/makerdao/gofer/pkg/graph"
)

type json struct {
	bufferedMarshaller *bufferedMarshaller
}

func newJSON(ndjson bool) *json {
	return &json{newBufferedMarshaller(ndjson, func(item interface{}) ([]marshalledItem, error) {
		if i, ok := item.([]marshalledItem); ok {
			b, err := encodingJson.Marshal(i)
			b = append(b, '\n')
			return []marshalledItem{b}, err
		}

		var err error
		var ret []marshalledItem

		switch i := item.(type) {
		case graph.AggregatorTick:
			err = jsonHandleTick(&ret, i)
		case graph.Aggregator:
			err = jsonHandleGraph(&ret, i)
		case map[graph.Pair][]string:
			err = jsonHandleOrigins(&ret, i)
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
func (j *json) Write(item interface{}) error {
	return j.bufferedMarshaller.Write(item)
}

// Close implements the Marshaller interface.
func (j *json) Close() error {
	return j.bufferedMarshaller.Close()
}

func jsonHandleTick(ret *[]marshalledItem, tick graph.AggregatorTick) error {
	b, err := encodingJson.Marshal(jsonTickFromAggregatorTick(tick))
	if err != nil {
		return err
	}

	b = append(b, '\n')
	*ret = append(*ret, b)
	return nil
}

func jsonHandleGraph(ret *[]marshalledItem, graph graph.Aggregator) error {
	b, err := encodingJson.Marshal(graph.Pair().String())
	if err != nil {
		return err
	}

	b = append(b, '\n')
	*ret = append(*ret, b)
	return nil
}

func jsonHandleOrigins(ret *[]marshalledItem, origins map[graph.Pair][]string) error {
	r := make(map[string][]string)
	for p, o := range origins {
		r[p.String()] = o
	}

	b, err := encodingJson.Marshal(r)
	if err != nil {
		return err
	}

	b = append(b, '\n')
	*ret = append(*ret, b)
	return nil
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
