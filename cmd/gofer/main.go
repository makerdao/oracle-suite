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

	"github.com/sirupsen/logrus"

	suite "github.com/makerdao/oracle-suite"
	"github.com/makerdao/oracle-suite/internal/config"
	"github.com/makerdao/oracle-suite/internal/gofer/marshal"
	"github.com/makerdao/oracle-suite/pkg/gofer"
	"github.com/makerdao/oracle-suite/pkg/gofer/rpc"
	"github.com/makerdao/oracle-suite/pkg/log"
	logLogrus "github.com/makerdao/oracle-suite/pkg/log/logrus"
)

// exitCode to be returned by the application.
var exitCode = 0

func main() {
	opts := options{
		Format:  formatTypeValue{format: marshal.NDJSON},
		Version: suite.Version,
	}

	rootCmd := NewRootCommand(&opts)
	rootCmd.AddCommand(
		NewPairsCmd(&opts),
		NewPricesCmd(&opts),
		NewAgentCmd(&opts),
	)

	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("Error: %s\n", err)
		if exitCode == 0 {
			os.Exit(1)
		}
	}
	os.Exit(exitCode)
}

func newLogger(opts *options) (log.Logger, error) {
	ll, err := logrus.ParseLevel(opts.LogVerbosity)
	if err != nil {
		return nil, err
	}

	lr := logrus.New()
	lr.SetLevel(ll)
	lr.SetFormatter(opts.LogFormat.Formatter())

	return logLogrus.New(lr), nil
}

func newGofer(opts *options, path string, logger log.Logger) (gofer.Gofer, error) {
	err := config.ParseFile(&opts.Config, path)
	if err != nil {
		return nil, err
	}
	return opts.Config.Configure(logger, opts.NoRPC)
}

func newAgent(opts *options, path string, logger log.Logger) (*rpc.Agent, error) {
	err := config.ParseFile(&opts.Config, path)
	if err != nil {
		return nil, err
	}
	return opts.Config.ConfigureRPCAgent(logger)
}
