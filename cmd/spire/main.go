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
	"io"
	"os"

	"github.com/sirupsen/logrus"

	suite "github.com/makerdao/oracle-suite"
	"github.com/makerdao/oracle-suite/internal/config"
	"github.com/makerdao/oracle-suite/pkg/log"
	logLogrus "github.com/makerdao/oracle-suite/pkg/log/logrus"
	"github.com/makerdao/oracle-suite/pkg/spire"
)

var (
	logger log.Logger
	client *spire.Client
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

func newAgent(opts *options, log log.Logger) (*spire.Agent, error) {
	err := config.ParseFile(&opts.Config, opts.ConfigFilePath)
	if err != nil {
		return nil, err
	}
	a, err := opts.Config.ConfigureAgent(Dependencies{
		Context: context.Background(),
		Logger:  log,
	})
	if err != nil {
		return nil, err
	}
	return a, nil
}

func newClient(opts *options) (*spire.Client, error) {
	err := config.ParseFile(&opts.Config, opts.ConfigFilePath)
	if err != nil {
		return nil, err
	}
	c, err := opts.Config.ConfigureClient()
	if err != nil {
		return nil, err
	}
	return c, nil
}

func readAll(r io.Reader) ([]byte, error) {
	b := make([]byte, 0, 512)
	for {
		if len(b) == cap(b) {
			// Add more capacity (let append pick how much).
			b = append(b, 0)[:len(b)]
		}
		n, err := r.Read(b[len(b):cap(b)])
		b = b[:len(b)+n]
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return b, err
		}
	}
}
