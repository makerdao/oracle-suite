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
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	"github.com/spf13/cobra"

	"github.com/makerdao/gofer/pkg/gofer"
	"github.com/makerdao/gofer/pkg/model"

	"github.com/makerdao/gofer/cmd/gofer/internal/command"
	"github.com/makerdao/gofer/cmd/gofer/internal/marshal"
)

type lib interface {
	Prices(pairs ...*model.Pair) (map[model.Pair]*model.PriceAggregate, error)
}

func newLib(path string) (lib, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	lib, err := gofer.ReadFile(absPath)
	if err != nil {
		return nil, err
	}

	return lib, nil
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

			var pairs []*model.Pair
			for _, pair := range args {
				var p *model.Pair
				if p, err = model.NewPairFromString(pair); err != nil {
					return err
				}
				pairs = append(pairs, model.NewPair(p.Base, p.Quote))
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

func prices(g lib, m *marshal.Marshal, pairs []*model.Pair) error {
	prices, err := g.Prices(pairs...)
	if err != nil {
		return err
	}

	for _, p := range sortPrices(prices) {
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

	return nil
}

func sortPrices(prices map[model.Pair]*model.PriceAggregate) []model.Pair {
	pairs := make([]model.Pair, 0)
	for k := range prices {
		pairs = append(pairs, k)
	}

	sort.SliceStable(pairs, func(i, j int) bool {
		return pairs[i].String() < pairs[j].String()
	})

	return pairs
}
