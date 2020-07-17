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

package aggregator

import (
	"fmt"
	"strings"
)

func (p *Pair) UnmarshalText(text []byte) error {
	s := string(text)
	ss := strings.Split(s, "/")
	// Pair should split into 2 parts, base and quote
	if len(ss) != 2 { //nolint:gomnd
		return fmt.Errorf("couldn't parse pair \"%s\"", s)
	}

	if len(ss[0]) == 0 {
		return fmt.Errorf("base asset name is empty")
	}

	if len(ss[1]) == 0 {
		return fmt.Errorf("quote asset name is empty")
	}

	p.Base = strings.ToUpper(ss[0])
	p.Quote = strings.ToUpper(ss[1])

	return nil
}

func (p Pair) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}
