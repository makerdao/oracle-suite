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

	"github.com/makerdao/gofer/pkg/gofer"
	"github.com/makerdao/gofer/pkg/model"

	"github.com/makerdao/gofer/cmd/gofer/internal/command"
	"github.com/makerdao/gofer/cmd/gofer/internal/marshal"
)

type lib interface {
	Exchanges(pairs ...*model.Pair) []*model.Exchange
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
		Use:   "exchanges [PAIR...]",
		Short: "List supported exchanges",
		Long: `
Lists exchanges that will be queried for all of the supported pairs
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

			var pairs []*model.Pair
			for _, pair := range args {
				var p *model.Pair
				if p, err = model.NewPairFromString(pair); err != nil {
					return err
				}
				pairs = append(pairs, model.NewPair(p.Base, p.Quote))
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

func exchanges(g lib, m *marshal.Marshal, pairs []*model.Pair) error {
	exchanges := g.Exchanges(pairs...)

	sort.SliceStable(exchanges, func(i, j int) bool {
		return exchanges[i].Name < exchanges[j].Name
	})

	var er error
	for _, e := range exchanges {
		err := m.Write(e, er)
		if err != nil {
			return err
		}
	}

	err := m.Close()
	if err != nil {
		return err
	}

	return nil
}
