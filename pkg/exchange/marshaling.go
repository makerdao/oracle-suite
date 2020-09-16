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

package exchange

import (
	"encoding/json"
	"strconv"
	"time"
)

//nolint:unused
type stringAsFloat float64

func (s *stringAsFloat) UnmarshalJSON(bytes []byte) error {
	var ss string
	if err := json.Unmarshal(bytes, &ss); err != nil {
		return err
	}
	f, err := strconv.ParseFloat(ss, 64)
	*s = stringAsFloat(f)
	return err
}

func (s *stringAsFloat) val() float64 {
	return float64(*s)
}

//nolint:unused
type stringAsUnixTimestamp time.Time

func (s *stringAsUnixTimestamp) UnmarshalJSON(bytes []byte) error {
	var ss string
	if err := json.Unmarshal(bytes, &ss); err != nil {
		return err
	}
	i, err := strconv.ParseInt(ss, 10, 64)
	*s = stringAsUnixTimestamp(time.Unix(i, 0))
	return err
}
func (s *stringAsUnixTimestamp) val() time.Time {
	return time.Time(*s)
}

//nolint:unused
type stringAsInt64 int64

func (s *stringAsInt64) UnmarshalJSON(bytes []byte) error {
	var ss string
	if err := json.Unmarshal(bytes, &ss); err != nil {
		return err
	}
	i, err := strconv.ParseInt(ss, 10, 64)
	*s = stringAsInt64(i)
	return err
}

func (s *stringAsInt64) val() int64 {
	return int64(*s)
}
