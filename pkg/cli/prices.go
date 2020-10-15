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

package cli

import (
	"errors"

	"github.com/makerdao/gofer/pkg/gofer"
)

func Prices(args []string, l gofer.PriceModels, m itemWriter) error {
	pairs, err := gofer.Pairs(l, args...)
	if err != nil {
		return err
	}

	ticks, err := l.Ticks(pairs...)
	if err != nil {
		return err
	}

	for _, t := range ticks {
		err = m.Write(t)
		if err != nil {
			return err
		}
	}

	for _, t := range ticks {
		if t.Error != nil {
			return errors.New("some of the prices were returned with an error")
		}
	}

	return nil
}
