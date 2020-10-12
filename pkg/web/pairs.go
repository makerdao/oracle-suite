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

package web

import (
	"io"
	"net/http"
	"sort"

	"github.com/makerdao/gofer/internal/marshal"
	"github.com/makerdao/gofer/pkg/graph"
)

func PairsHandler(l graph.PriceModels) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var graphs []graph.Aggregator
		for _, g := range l {
			graphs = append(graphs, g)
		}

		sort.SliceStable(graphs, func(i, j int) bool {
			return graphs[i].Pair().String() < graphs[j].Pair().String()
		})

		m, _ := marshal.NewMarshal(marshal.JSON)
		for _, g := range graphs {
			if err := m.Write(g); err != nil {
				BadRequest(w, err)
				return
			}
		}

		if err := m.Close(); err != nil {
			BadRequest(w, err)
			return
		}

		if _, err := io.Copy(w, m); err != nil {
			BadRequest(w, err)
			return
		}

		AsJSON(w)
	}
}
