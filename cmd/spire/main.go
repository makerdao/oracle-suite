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
	"io"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"

	"github.com/makerdao/gofer/pkg/log"
	logLogrus "github.com/makerdao/gofer/pkg/log/logrus"
	"github.com/makerdao/gofer/pkg/spire"
	"github.com/makerdao/gofer/pkg/spire/config"
	configJSON "github.com/makerdao/gofer/pkg/spire/config/json"
)

var (
	logger log.Logger
	client *spire.Client
)

func main() {
	var opts options
	rootCmd := NewRootCommand(&opts)

	if err := rootCmd.Execute(); err != nil {
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

func newServer(opts *options, log log.Logger) (*spire.Agent, error) {
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

	s, err := opts.Config.ConfigureServer(config.Dependencies{
		Context: context.Background(),
		Logger:  log,
	})
	if err != nil {
		return nil, err
	}

	return s, nil
}

func newClient(opts *options, log log.Logger) (*spire.Client, error) {
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

	c, err := opts.Config.ConfigureClient(config.Dependencies{
		Context: context.Background(),
		Logger:  log,
	})
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
