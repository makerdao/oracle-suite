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

type pairsLister interface {
	Graphs() map[graph.Pair]graph.Aggregator
}

func Pairs(l pairsLister, m ReadWriteCloser) error {
	var err error

	var graphs []graph.Aggregator
	for _, g := range l.Graphs() {
		graphs = append(graphs, g)
	}

	sort.SliceStable(graphs, func(i, j int) bool {
		return graphs[i].Pair().String() < graphs[j].Pair().String()
	})

	for _, g := range graphs {
		err = m.Write(g, nil)
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
