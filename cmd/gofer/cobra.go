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
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/spf13/cobra"

	"github.com/makerdao/gofer/internal/gofer/cli"
	"github.com/makerdao/gofer/internal/gofer/marshal"
	"github.com/makerdao/gofer/internal/gofer/web"
	"github.com/makerdao/gofer/pkg/gofer"
	configJSON "github.com/makerdao/gofer/pkg/gofer/config/json"
	"github.com/makerdao/gofer/pkg/gofer/graph"
	"github.com/makerdao/gofer/pkg/gofer/origins"
)

func newGofer(opts *options, path string) (*gofer.Gofer, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	err = configJSON.ParseJSONFile(&opts.Config, absPath)
	if err != nil {
		return nil, err
	}

	g, err := opts.Config.BuildGraphs()
	if err != nil {
		return nil, err
	}

	return gofer.NewGofer(g, graph.NewFeeder(origins.DefaultSet())), nil
}

// asyncCopy asynchronously copies from src to dst using the io.Copy.
// The returned function will block the current goroutine until
// the io.Copy finished.
func asyncCopy(dst io.Writer, src io.Reader) func() {
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		_, err := io.Copy(dst, src)
		wg.Done()

		if err == io.EOF {
			return
		}
		if err != nil {
			panic(err.Error())
		}
	}()

	return func() {
		wg.Wait()
	}
}

func NewPairsCmd(o *options) *cobra.Command {
	return &cobra.Command{
		Use:     "pairs",
		Aliases: []string{"pair"},
		Args:    cobra.NoArgs,
		Short:   "List all supported pairs",
		Long:    `List all supported asset pairs.`,
		RunE: func(_ *cobra.Command, args []string) error {
			m, err := marshal.NewMarshal(o.OutputFormat.format)
			if err != nil {
				return err
			}

			wait := asyncCopy(os.Stdout, m)
			defer func() {
				_ = m.Close()
				wait()
			}()

			absPath, err := filepath.Abs(o.ConfigFilePath)
			if err != nil {
				return err
			}

			g, err := newGofer(o, absPath)
			if err != nil {
				return err
			}

			err = cli.Pairs(g, m)
			if err != nil {
				return err
			}

			return nil
		},
	}
}

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

			wait := asyncCopy(os.Stdout, m)
			defer func() {
				_ = m.Close()
				wait()
			}()

			absPath, err := filepath.Abs(o.ConfigFilePath)
			if err != nil {
				return err
			}

			g, err := newGofer(o, absPath)
			if err != nil {
				return err
			}

			err = cli.Origins(args, g, m)
			if err != nil {
				return err
			}

			return nil
		},
	}
}

func NewPricesCmd(o *options) *cobra.Command {
	return &cobra.Command{
		Use:     "prices [PAIR...]",
		Aliases: []string{"price"},
		Args:    cobra.MinimumNArgs(0),
		Short:   "Return price for given PAIRs",
		Long:    `Print the price of given PAIRs`,
		RunE: func(_ *cobra.Command, args []string) error {
			m, err := marshal.NewMarshal(o.OutputFormat.format)
			if err != nil {
				return err
			}

			wait := asyncCopy(os.Stdout, m)
			defer func() {
				_ = m.Close()
				wait()
			}()

			absPath, err := filepath.Abs(o.ConfigFilePath)
			if err != nil {
				return err
			}

			g, err := newGofer(o, absPath)
			if err != nil {
				return err
			}

			if err := g.Feed(g.Pairs()...); err != nil {
				return err
			}

			err = cli.Prices(args, g, m)
			if err != nil {
				return err
			}

			return nil
		},
	}
}

func NewServerCmd(o *options) *cobra.Command {
	return &cobra.Command{
		Use:   "server",
		Args:  cobra.ExactArgs(0),
		Short: "",
		Long:  ``,
		RunE: func(_ *cobra.Command, _ []string) error {
			absPath, err := filepath.Abs(o.ConfigFilePath)
			if err != nil {
				return err
			}

			g, err := newGofer(o, absPath)
			if err != nil {
				return err
			}

			http.HandleFunc("/v1/pairs/", web.PairsHandler(g))
			http.HandleFunc("/v1/origins/", web.OriginsHandler(g))
			http.HandleFunc("/v1/prices/", web.PricesHandler(g))

			// TODO: log errors
			_ = g.StartFeeder(g.Pairs()...)

			return web.StartServer(":8080")
		},
	}
}

func NewRootCommand(opts *options) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:     "gofer",
		Version: "DEV",
		Short:   "Tool for providing reliable data in the blockchain ecosystem",
		Long: `
Gofer is a CLI interface for the Gofer Go Library.

It is a tool that allows for easy data retrieval from various sources
with aggregates that increase reliability in the DeFi environment.`,
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	rootCmd.PersistentFlags().StringVarP(&opts.ConfigFilePath, "config", "c", "./gofer.json", "config file")
	rootCmd.PersistentFlags().VarP(&opts.OutputFormat, "format", "f", "output format")

	return rootCmd
}
