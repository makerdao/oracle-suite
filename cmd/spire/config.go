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

	ethereumConfig "github.com/makerdao/oracle-suite/internal/config/ethereum"
	feedsConfig "github.com/makerdao/oracle-suite/internal/config/feeds"
	spireConfig "github.com/makerdao/oracle-suite/internal/config/spire"
	transportConfig "github.com/makerdao/oracle-suite/internal/config/transport"
	"github.com/makerdao/oracle-suite/pkg/datastore"
	"github.com/makerdao/oracle-suite/pkg/log"
	"github.com/makerdao/oracle-suite/pkg/spire"
	"github.com/makerdao/oracle-suite/pkg/transport"
)

type Config struct {
	Transport transportConfig.Transport `json:"transport"`
	Ethereum  ethereumConfig.Ethereum   `json:"ethereum"`
	Spire     spireConfig.Spire         `json:"spire"`
	Feeds     feedsConfig.Feeds         `json:"feeds"`
}

type ClientDependencies struct {
	Context context.Context
}

type AgentDependencies struct {
	Context context.Context
	Logger  log.Logger
}

func (c *Config) ConfigureClient(d ClientDependencies) (*spire.Client, error) {
	sig, err := c.Ethereum.ConfigureSigner()
	if err != nil {
		return nil, err
	}
	cli, err := c.Spire.ConfigureClient(spireConfig.ClientDependencies{
		Context: d.Context,
		Signer:  sig,
	})
	if err != nil {
		return nil, err
	}
	return cli, nil
}

func (c *Config) ConfigureAgent(d AgentDependencies) (transport.Transport, datastore.Datastore, *spire.Agent, error) {
	sig, err := c.Ethereum.ConfigureSigner()
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
	dat, err := c.Spire.ConfigureDatastore(spireConfig.DatastoreDependencies{
		Context:   d.Context,
		Signer:    sig,
		Transport: tra,
		Feeds:     fed,
		Logger:    d.Logger,
	})
	if err != nil {
		return nil, nil, nil, err
	}
	age, err := c.Spire.ConfigureAgent(spireConfig.AgentDependencies{
		Context:   d.Context,
		Signer:    sig,
		Transport: tra,
		Datastore: dat,
		Feeds:     fed,
		Logger:    d.Logger,
	})
	if err != nil {
		return nil, nil, nil, err
	}
	return tra, dat, age, nil
}
