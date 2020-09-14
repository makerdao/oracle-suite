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
	"io"

	"github.com/makerdao/gofer/pkg/graph"
)

type ReadWriteCloser interface {
	io.ReadCloser
	Write(item interface{}, err error) error
}

type pricer interface {
	Ticks(pairs ...graph.Pair) ([]graph.IndirectTick, error)
	Pairs() []graph.Pair
}

func Price(args []string, l pricer, m ReadWriteCloser) error {
	var pairs []graph.Pair

	if len(args) > 0 {
		for _, pair := range args {
			p, err := graph.NewPair(pair)
			if err != nil {
				return err
			}
			pairs = append(pairs, p)
		}
	} else {
		pairs = l.Pairs()
	}

	ticks, err := l.Ticks(pairs...)
	if err != nil {
		return err
	}

	for _, t := range ticks {
		err = m.Write(t, nil)
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
