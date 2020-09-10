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

package price

import (
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/makerdao/gofer/cmd/gofer/internal/command"
	"github.com/makerdao/gofer/cmd/gofer/internal/marshal"
	"github.com/makerdao/gofer/pkg/config"
	"github.com/makerdao/gofer/pkg/exchange"
	"github.com/makerdao/gofer/pkg/gofer"
	"github.com/makerdao/gofer/pkg/graph"
)

type lib interface {
	Ticks(pairs ...graph.Pair) ([]graph.IndirectTick, error)
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
		Use:   "price PAIR [PAIR...]",
		Args:  cobra.MinimumNArgs(1),
		Short: "Return price for given PAIRs",
		Long:  `Print the price of given PAIRs`,
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

			err = prices(l, m, pairs)
			if err != nil {
				return err
			}

			_, err = io.Copy(os.Stdout, m)
			if err != nil {
				return err
			}

			return nil
		},
	}
}

func prices(g lib, m *marshal.Marshal, pairs []graph.Pair) error {
	var err error

	ticks, err := g.Ticks(pairs...)
	if err != nil {
		return err
	}

	for _, t := range ticks {
		err := m.Write(t, nil)
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

