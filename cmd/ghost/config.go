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
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/makerdao/oracle-suite/internal/config"
	ethereumConfig "github.com/makerdao/oracle-suite/internal/config/ethereum"
	feedsConfig "github.com/makerdao/oracle-suite/internal/config/feeds"
	ghostConfig "github.com/makerdao/oracle-suite/internal/config/ghost"
	goferConfig "github.com/makerdao/oracle-suite/internal/config/gofer"
	transportConfig "github.com/makerdao/oracle-suite/internal/config/transport"
	"github.com/makerdao/oracle-suite/pkg/ethereum"
	"github.com/makerdao/oracle-suite/pkg/ghost"
	logLogrus "github.com/makerdao/oracle-suite/pkg/log/logrus"
	"github.com/makerdao/oracle-suite/pkg/transport"

	"github.com/makerdao/oracle-suite/pkg/log"
)

type Config struct {
	Gofer     goferConfig.Gofer         `json:"gofer"`
	Ethereum  ethereumConfig.Ethereum   `json:"ethereum"`
	Transport transportConfig.Transport `json:"transport"`
	Ghost     ghostConfig.Ghost         `json:"ghost"`
	Feeds     feedsConfig.Feeds         `json:"feeds"`
}

type Dependencies struct {
	Context context.Context
	Logger  log.Logger
}

func (c *Config) Configure(d Dependencies) (transport.Transport, *ghost.Ghost, error) {
	gof, err := c.Gofer.ConfigureGofer(d.Context, d.Logger, true)
	if err != nil {
		return nil, nil, err
	}
	sig, err := c.Ethereum.ConfigureSigner()
	if err != nil {
		return nil, nil, err
	}
	if sig.Address() == ethereum.EmptyAddress {
		return nil, nil, errors.New("ethereum account must be configured")
	}
	fed, err := c.Feeds.Addresses()
	if err != nil {
		return nil, nil, err
	}
	tra, err := c.Transport.Configure(transportConfig.Dependencies{
		Context: d.Context,
		Signer:  sig,
		Feeds:   fed,
		Logger:  d.Logger,
	})
	if err != nil {
		return nil, nil, err
	}
	gho, err := c.Ghost.Configure(ghostConfig.Dependencies{
		Context:   d.Context,
		Gofer:     gof,
		Signer:    sig,
		Transport: tra,
		Logger:    d.Logger,
	})
	if err != nil {
		return nil, nil, err
	}
	return tra, gho, nil
}

type Services struct {
	ctxCancel context.CancelFunc
	Transport transport.Transport
	Ghost     *ghost.Ghost
}

func PrepareServices(ctx context.Context, opts *options) (*Services, error) {
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
		return nil, fmt.Errorf("failed to parse configuration file: %w", err)
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
		return nil, fmt.Errorf("failed to load Ghost configuration: %w", err)
	}

	return &Services{
		ctxCancel: ctxCancel,
		Transport: tra,
		Ghost:     gho,
	}, nil
}

func (s *Services) Start() error {
	var err error
	if err = s.Transport.Start(); err != nil {
		return err
	}
	if err = s.Ghost.Start(); err != nil {
		return err
	}
	return nil
}

func (s *Services) CancelAndWait() {
	s.ctxCancel()
	s.Transport.Wait()
	s.Ghost.Wait()
}
