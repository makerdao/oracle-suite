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
	_ "embed"
	"os"
	"path/filepath"

	"github.com/makerdao/oracle-suite/cmd/spammen/internal"

	"github.com/makerdao/oracle-suite/pkg/spire/config"

	"github.com/sirupsen/logrus"

	suite "github.com/makerdao/oracle-suite"
	"github.com/makerdao/oracle-suite/pkg/log"
	logLogrus "github.com/makerdao/oracle-suite/pkg/log/logrus"
	configJSON "github.com/makerdao/oracle-suite/pkg/spire/config/json"
)

var (
	logger log.Logger
	client *internal.Spamman
)

func main() {
	opts := options{Version: suite.Version}
	rootCmd := NewRootCommand(&opts)

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

func newSpamman(opts *options, log log.Logger) (*internal.Spamman, error) {
	if opts.ConfigPath != "" {
		absPath, err := filepath.Abs(opts.ConfigPath)
		if err != nil {
			return nil, err
		}

		err = configJSON.ParseJSONFile(&opts.Config, absPath)
		if err != nil {
			return nil, err
		}
	}

	c, err := opts.Config.Configure(config.Dependencies{
		Context: context.Background(),
		Logger:  log,
	})
	if err != nil {
		return nil, err
	}

	s := internal.NewSpamman(internal.Config{
		MessageRate:      opts.MessageRate,
		ValidMessages:    opts.ValidMessages,
		InvalidSignature: opts.InvalidSignature,
		Pairs:            opts.Config.Pairs,
		Instance:         c,
	})
	return s, nil
}
