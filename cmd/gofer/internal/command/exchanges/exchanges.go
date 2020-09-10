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

package exchanges

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sort"

	"github.com/spf13/cobra"

	"github.com/makerdao/gofer/cmd/gofer/internal/command"
	"github.com/makerdao/gofer/cmd/gofer/internal/marshal"
	"github.com/makerdao/gofer/pkg/config"
	"github.com/makerdao/gofer/pkg/exchange"
	"github.com/makerdao/gofer/pkg/gofer"
	"github.com/makerdao/gofer/pkg/graph"
)

type lib interface {
	Exchanges(pairs ...graph.Pair) (map[graph.Pair][]string, error)
}

func newLib(path string) (lib, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	j, err := config.ParseJSONFile(absPath)
	if err != nil {
		return nil, err
	}

	g, err := j.BuildGraphs()
	if err != nil {
		return nil, err
	}

	return gofer.NewGofer(g, graph.NewIngestor(exchange.DefaultSet(), 10)), nil
}

func New(opts *command.Options) *cobra.Command {
	return &cobra.Command{
		Use:   "exchanges [PAIR...]",
		Short: "List supported exchanges",
		Long: `Lists exchanges that will be queried for all of the supported pairs
or a subset of those, if at least one PAIR is provided.`,
		RunE: func(c *cobra.Command, args []string) error {
			m, err := marshal.NewMarshal(opts.OutputFormat.Format)
			if err != nil {
				return err
			}

			l, err := newLib(opts.ConfigFilePath)
			if err != nil {
				return err
			}

			var pairs []graph.Pair
			for _, pair := range args {
				var p graph.Pair
				if p, err = graph.NewPair(pair); err != nil {
					return err
				}
				pairs = append(pairs, p)
			}

			err = exchanges(l, m, pairs)
			if err != nil {
				return err
			}

			b, err := ioutil.ReadAll(m)
			if err != nil {
				return err
			}

			fmt.Print(string(b))

			return nil
		},
	}
}

func exchanges(g lib, m *marshal.Marshal, pairs []graph.Pair) error {
	var err error

	exchanges, err := g.Exchanges(pairs...)
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
