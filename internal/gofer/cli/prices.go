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

	"github.com/makerdao/gofer/pkg/gofer/feeder"
	"github.com/makerdao/gofer/pkg/gofer/graph"
)

type pricer interface {
	Feed(pairs ...graph.Pair) (feeder.Warnings, error)
	Ticks(pairs ...graph.Pair) ([]graph.AggregatorTick, error)
	Pairs() []graph.Pair
}

func Prices(args []string, l pricer, m itemWriter) error {
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

	_, err := l.Feed(pairs...)
	if err != nil {
		return err
	}

	ticks, err := l.Ticks(pairs...)
	if err != nil {
		return err
	}

	sort.SliceStable(ticks, func(i, j int) bool {
		return ticks[i].Pair.String() < ticks[j].Pair.String()
	})

	for _, t := range ticks {
		err = m.Write(t)
		if err != nil {
			return err
		}
	}

	return nil
}
