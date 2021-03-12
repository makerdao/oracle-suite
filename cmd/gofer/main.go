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
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"

	"github.com/makerdao/gofer/internal/gofer/marshal"
	"github.com/makerdao/gofer/pkg/gofer"
	configJSON "github.com/makerdao/gofer/pkg/gofer/config/json"
	"github.com/makerdao/gofer/pkg/gofer/rpc"
	"github.com/makerdao/gofer/pkg/log"
	logLogrus "github.com/makerdao/gofer/pkg/log/logrus"
)

// errSilent is used to return an non-zero status code without an error
// message.
type errSilent struct{}

func (e errSilent) Error() string { return "" }

func main() {
	opts := options{
		OutputFormat: formatTypeValue{format: marshal.NDJSON},
	}

	rootCmd := NewRootCommand(&opts)
	rootCmd.AddCommand(
		NewPairsCmd(&opts),
		NewPricesCmd(&opts),
		NewAgentCmd(&opts),
	)

	if err := rootCmd.Execute(); err != nil {
		if _, ok := err.(errSilent); !ok {
			fmt.Printf("Error: %s\n", err)
		}
		os.Exit(1)
	}
}

func newLogger(level string) (log.Logger, error) {
	ll, err := logrus.ParseLevel(level)
	if err != nil {
		return nil, err
	}

	lr := logrus.New()
	lr.SetLevel(ll)

	return logLogrus.New(lr), nil
}

func newGofer(opts *options, path string, logger log.Logger) (gofer.Gofer, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	err = configJSON.ParseJSONFile(&opts.Config, absPath)
	if err != nil {
		return nil, err
	}

	var gof gofer.Gofer
	if opts.Config.RPC.Address == "" || opts.NoRPC {
		gof, err = opts.Config.ConfigureGofer(logger)
		if err != nil {
			return nil, err
		}
	} else {
		gof, err = opts.Config.ConfigureRPCClient(logger)
		if err != nil {
			return nil, err
		}
	}

	return gof, nil
}

func newServer(opts *options, path string, logger log.Logger) (*rpc.Agent, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	err = configJSON.ParseJSONFile(&opts.Config, absPath)
	if err != nil {
		return nil, err
	}

	return opts.Config.ConfigureRPCServer(logger)
}
