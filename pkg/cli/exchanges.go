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
	"sort"

	"github.com/makerdao/gofer/pkg/graph"
)

type exchangeLister interface {
	Exchanges(pairs ...graph.Pair) (map[graph.Pair][]string, error)
}

func Exchanges(args []string, l exchangeLister, m ReadWriteCloser) error {
	var err error

	var pairs []graph.Pair
	for _, pair := range args {
		p, err := graph.NewPair(pair)
		if err != nil {
			return err
		}
		pairs = append(pairs, p)
	}

	exchanges, err := l.Exchanges(pairs...)
	if err != nil {
		return err
	}

	var list []string
	for _, e := range exchanges {
		list = append(list, e...)
	}

	sort.SliceStable(list, func(i, j int) bool {
		return list[i] < list[j]
	})

	for _, name := range list {
		err := m.Write(name, nil)
		if err != nil {
			return err
		}
	}

	err = m.Close()
	if err != nil {
		return err
	}

	return nil
}
