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

package pairs

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sort"

	"github.com/spf13/cobra"

	"github.com/makerdao/gofer/pkg/config"
	"github.com/makerdao/gofer/pkg/exchange"
	"github.com/makerdao/gofer/pkg/gofer"
	"github.com/makerdao/gofer/pkg/graph"

	"github.com/makerdao/gofer/cmd/gofer/internal/command"
	"github.com/makerdao/gofer/cmd/gofer/internal/marshal"
)

type lib interface {
	Graphs() map[graph.Pair]graph.Aggregator
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
		Use:   "pairs",
		Args:  cobra.NoArgs,
		Short: "List all supported pairs",
		Long:  `List all supported asset pairs.`,
		RunE: func(c *cobra.Command, args []string) error {
			m, err := marshal.NewMarshal(opts.OutputFormat.Format)
			if err != nil {
				return err
			}

			l, err := newLib(opts.ConfigFilePath)
			if err != nil {
				return err
			}

			err = pairs(l, m)
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

func pairs(l lib, m *marshal.Marshal) error {
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
