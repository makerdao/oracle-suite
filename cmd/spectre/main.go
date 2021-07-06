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
	"context"
	"os"

	"github.com/sirupsen/logrus"

	"github.com/makerdao/oracle-suite/internal/config"
	"github.com/makerdao/oracle-suite/pkg/log"
	logLogrus "github.com/makerdao/oracle-suite/pkg/log/logrus"
	"github.com/makerdao/oracle-suite/pkg/spectre"
)

func main() {
	var opts options
	rootCmd := NewRootCommand(&opts)

	rootCmd.AddCommand(
		NewRunCmd(&opts),
	)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
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

func newSpectre(opts *options, logger log.Logger) (*spectre.Spectre, error) {
	err := config.ParseFile(&opts.Config, opts.ConfigFilePath)
	if err != nil {
		return nil, err
	}
	return opts.Config.Configure(Dependencies{
		Context: context.Background(),
		Logger:  logger,
	})
}
