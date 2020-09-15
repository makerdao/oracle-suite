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

package graph

//TODO: This is to be moved to a separate marshaling package
import (
	"encoding/json"
	"time"
)

func (p OriginTick) MarshalJSON() ([]byte, error) {
	var errStr string
	ts := &p.Timestamp
	if p.Error != nil {
		errStr = p.Error.Error()
		ts = nil
	}

	return json.Marshal(struct {
		Type   string `json:"type,omitempty"`
		Origin string `json:"origin,omitempty"`
		// <Tick>
		Base      string     `json:"base,omitempty"`
		Quote     string     `json:"quote,omitempty"`
		Price     float64    `json:"price,omitempty"`
		Bid       float64    `json:"bid,omitempty"`
		Ask       float64    `json:"ask,omitempty"`
		Volume24h float64    `json:"vol24h,omitempty"`
		Timestamp *time.Time `json:"ts,omitempty"`
		// </Tick>
		Error string `json:"error,omitempty"`
	}{
		Type:   "origin",
		Origin: p.Origin,
		// <Tick>
		Base:      p.Pair.Base,
		Quote:     p.Pair.Quote,
		Price:     p.Price,
		Bid:       p.Bid,
		Ask:       p.Ask,
		Volume24h: p.Volume24h,
		Timestamp: ts,
		// </Tick>
		Error: errStr,
	})
}

func (p AggregatorTick) MarshalJSON() ([]byte, error) {
	var errStr string
	ts := &p.Timestamp
	if p.Error != nil {
		errStr = p.Error.Error()
		ts = nil
	}

	var ticks []interface{}
	for _, v := range p.OriginTicks {
		ticks = append(ticks, v)
	}
	for _, v := range p.AggregatorTicks {
		ticks = append(ticks, v)
	}

	return json.Marshal(struct {
		Type   string `json:"type,omitempty"`
		Method string `json:"method,omitempty"`
		// <Tick>
		Base      string     `json:"base,omitempty"`
		Quote     string     `json:"quote,omitempty"`
		Price     float64    `json:"price,omitempty"`
		Bid       float64    `json:"bid,omitempty"`
		Ask       float64    `json:"ask,omitempty"`
		Volume24h float64    `json:"vol24h,omitempty"`
		Timestamp *time.Time `json:"ts,omitempty"`
		// </Tick>
		Ticks []interface{} `json:"ticks,omitempty"`
		Error string        `json:"error,omitempty"`
	}{
		Type:   "aggregate",
		Method: p.Method,
		// <Tick>
		Base:      p.Pair.Base,
		Quote:     p.Pair.Quote,
		Price:     p.Price,
		Bid:       p.Bid,
		Ask:       p.Ask,
		Volume24h: p.Volume24h,
		Timestamp: ts,
		// </Tick>
		Ticks: ticks,
		Error: errStr,
	})
}
