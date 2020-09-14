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

import "encoding/json"

func (p OriginTick) MarshalJSON() ([]byte, error) {
	var errStr string
	if p.Error != nil {
		errStr = p.Error.Error()
	}

	return json.Marshal(struct {
		Tick
		Origin string
		Error  string
	}{
		Tick:   p.Tick,
		Origin: p.Origin,
		Error:  errStr,
	})
}

func (p IndirectTick) MarshalJSON() ([]byte, error) {
	var errStr string
	if p.Error != nil {
		errStr = p.Error.Error()
	}

	return json.Marshal(struct {
		Tick
		OriginTicks   []OriginTick
		IndirectTicks []IndirectTick
		Error         string
	}{
		Tick:          p.Tick,
		OriginTicks:   p.OriginTicks,
		IndirectTicks: p.IndirectTicks,
		Error:         errStr,
	})
}
