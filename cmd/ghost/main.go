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
	"github.com/makerdao/oracle-suite/pkg/ghost"
	logLogrus "github.com/makerdao/oracle-suite/pkg/log/logrus"
	"github.com/makerdao/oracle-suite/pkg/transport"
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

type services struct {
	ctxCancel context.CancelFunc
	transport transport.Transport
	ghost     *ghost.Ghost
}

func newServices(ctx context.Context, opts *options) (*services, error) {
	var err error
	ctx, ctxCancel := context.WithCancel(ctx)
	defer func() {
		if err != nil {
			ctxCancel()
		}
	}()

	// Load config file:
	err = config.ParseFile(&opts.Config, opts.ConfigFilePath)
	if err != nil {
		return nil, err
	}

	// Logger:
	ll, err := logrus.ParseLevel(opts.LogVerbosity)
	if err != nil {
		return nil, err
	}
	lr := logrus.New()
	lr.SetLevel(ll)
	lr.SetFormatter(opts.LogFormat.Formatter())
	logger := logLogrus.New(lr)

	// Services:
	tra, gho, err := opts.Config.Configure(Dependencies{
		Context: ctx,
		Logger:  logger,
	})
	if err != nil {
		return nil, err
	}

	return &services{
		ctxCancel: ctxCancel,
		transport: tra,
		ghost:     gho,
	}, nil
}

func (s *services) start() error {
	var err error
	if err = s.transport.Start(); err != nil {
		return err
	}
	if err = s.ghost.Start(); err != nil {
		return err
	}
	return nil
}

func (s *services) cancelAndWait() {
	s.ctxCancel()
	s.transport.Wait()
	s.ghost.Wait()
}
