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
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/makerdao/gofer/pkg/model"
)

type ReadWriteCloser interface {
	io.ReadCloser
	Write(item interface{}, err error) error
}

type pricer interface {
	Prices(pairs ...*model.Pair) (map[model.Pair]*model.PriceAggregate, error)
}

func Price(args []string, l pricer, m ReadWriteCloser) error {
	var pairs []*model.Pair
	for _, pair := range args {
		p, err := model.NewPairFromString(pair)
		if err != nil {
			return err
		}
		pairs = append(pairs, model.NewPair(p.Base, p.Quote))
	}

	prices, err := l.Prices(pairs...)
	if err != nil {
		return err
	}

	keys := make([]model.Pair, 0)
	for k := range prices {
		keys = append(keys, k)
	}

	sort.SliceStable(keys, func(i, j int) bool {
		return keys[i].String() < keys[j].String()
	})

	for _, p := range keys {
		var e error
		if prices[p] == nil {
			e = fmt.Errorf("no price aggregate available for %s available", p.String())
		} else if prices[p].Price == 0 {
			e = fmt.Errorf("invalid price for %s", p.String())
		}

		err = m.Write(prices[p], e)
		if err != nil {
			return err
		}
	}

	err = m.Close()
	if err != nil {
		return err
	}

	_, err = io.Copy(os.Stdout, m)
	if err != nil {
		return err
	}

	return nil
}
