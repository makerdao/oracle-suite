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

package main

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/makerdao/gofer/internal/gofer/cli"
	"github.com/makerdao/gofer/internal/gofer/marshal"
)

func NewOriginsCmd(o *options) *cobra.Command {
	return &cobra.Command{
		Use:     "origins [PAIR...]",
		Aliases: []string{"origin", "exchanges", "exchange", "sources", "source"},
		Short:   "List supported origins",
		Long: `Lists origins that will be queried for all of the supported pairs
or a subset of those, if at least one PAIR is provided.`,
		RunE: func(_ *cobra.Command, args []string) error {
			m, err := marshal.NewMarshal(o.OutputFormat.format)
			if err != nil {
				return err
			}

			absPath, err := filepath.Abs(o.ConfigFilePath)
			if err != nil {
				return err
			}

			l, err := newLogger(o.LogVerbosity)
			if err != nil {
				return err
			}

			g, err := newGofer(o, absPath, l)
			if err != nil {
				return err
			}

			err = cli.Origins(args, g, m)
			if err != nil {
				return err
			}

			bts, err := m.Bytes()
			if err != nil {
				return err
			}
			fmt.Print(string(bts))

			return nil
		},
	}
}
