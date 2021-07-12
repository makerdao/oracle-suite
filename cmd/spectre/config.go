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

	"github.com/sirupsen/logrus"

	"github.com/makerdao/oracle-suite/internal/config"
	ethereumConfig "github.com/makerdao/oracle-suite/internal/config/ethereum"
	feedsConfig "github.com/makerdao/oracle-suite/internal/config/feeds"
	spectreConfig "github.com/makerdao/oracle-suite/internal/config/spectre"
	transportConfig "github.com/makerdao/oracle-suite/internal/config/transport"
	"github.com/makerdao/oracle-suite/pkg/datastore"
	"github.com/makerdao/oracle-suite/pkg/log"
	logLogrus "github.com/makerdao/oracle-suite/pkg/log/logrus"
	"github.com/makerdao/oracle-suite/pkg/spectre"
	"github.com/makerdao/oracle-suite/pkg/transport"
)

type Config struct {
	Transport transportConfig.Transport `json:"transport"`
	Ethereum  ethereumConfig.Ethereum   `json:"ethereum"`
	Spectre   spectreConfig.Spectre     `json:"spectre"`
	Feeds     feedsConfig.Feeds         `json:"feeds"`
}

type Dependencies struct {
	Context context.Context
	Logger  log.Logger
}

func (c *Config) Configure(d Dependencies) (transport.Transport, datastore.Datastore, *spectre.Spectre, error) {
	sig, err := c.Ethereum.ConfigureSigner()
	if err != nil {
		return nil, nil, nil, err
	}
	cli, err := c.Ethereum.ConfigureEthereumClient(sig)
	if err != nil {
		return nil, nil, nil, err
	}
	fed, err := c.Feeds.Addresses()
	if err != nil {
		return nil, nil, nil, err
	}
	tra, err := c.Transport.Configure(transportConfig.Dependencies{
		Context: d.Context,
		Signer:  sig,
		Feeds:   fed,
		Logger:  d.Logger,
	})
	if err != nil {
		return nil, nil, nil, err
	}
	dat, err := c.Spectre.ConfigureDatastore(spectreConfig.DatastoreDependencies{
		Signer:    sig,
		Transport: tra,
		Feeds:     fed,
		Logger:    d.Logger,
	})
	if err != nil {
		return nil, nil, nil, err
	}
	spe, err := c.Spectre.ConfigureSpectre(spectreConfig.Dependencies{
		Signer:         sig,
		Datastore:      dat,
		EthereumClient: cli,
		Logger:         d.Logger,
	})
	if err != nil {
		return nil, nil, nil, err
	}
	return tra, dat, spe, nil
}

type Services struct {
	ctxCancel context.CancelFunc
	Transport transport.Transport
	Datastore datastore.Datastore
	Spectre   *spectre.Spectre
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
	tra, dat, spe, err := opts.Config.Configure(Dependencies{
		Context: ctx,
		Logger:  logger,
	})
	if err != nil {
		return nil, err
	}

	return &Services{
		ctxCancel: ctxCancel,
		Transport: tra,
		Datastore: dat,
		Spectre:   spe,
	}, nil
}

func (s *Services) Start() error {
	var err error
	if err = s.Transport.Start(); err != nil {
		return err
	}
	if err = s.Datastore.Start(); err != nil {
		return err
	}
	if err = s.Spectre.Start(); err != nil {
		return err
	}
	return nil
}

func (s *Services) CancelAndWait() {
	s.ctxCancel()
	s.Transport.Wait()
	s.Datastore.Wait()
	s.Spectre.Wait()
}
