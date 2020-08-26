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
	"sort"

	"github.com/spf13/cobra"

	"github.com/makerdao/gofer/aggregator"

	"github.com/makerdao/gofer/cmd/gofer/internal/command"
	"github.com/makerdao/gofer/cmd/gofer/internal/config"
	"github.com/makerdao/gofer/cmd/gofer/internal/marshal"
)

func New(opts *command.Options) *cobra.Command {
	return &cobra.Command{
		Use:   "pairs",
		Args:  cobra.NoArgs,
		Short: "List all supported pairs",
		Long: `
List all supported asset pairs.`,
		RunE: func(c *cobra.Command, args []string) error {
			m, err := marshal.NewMarshal(opts.OutputFormat.Format)
			if err != nil {
				return err
			}

			conf, err := config.ReadConfig(opts.ConfigFilePath)
			if err != nil {
				return err
			}

			err = pairs(conf.Aggregator.Parameters.PriceModels, m)
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

func pairs(s aggregator.PriceModelMap, m *marshal.Marshal) error {
	var err error

	for _, p := range sortPriceModels(s) {
		err = m.Write(aggregator.PriceModelMap{p: s[p]}, nil)
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

func sortPriceModels(models aggregator.PriceModelMap) []aggregator.Pair {
	pairs := make([]aggregator.Pair, 0)
	for k := range models {
		pairs = append(pairs, k)
	}

	sort.SliceStable(pairs, func(i, j int) bool {
		return pairs[i].String() < pairs[j].String()
	})

	return pairs
}
