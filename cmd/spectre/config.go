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
	spectreConfig "github.com/makerdao/oracle-suite/internal/config/spectre"
	transportConfig "github.com/makerdao/oracle-suite/internal/config/transport"
	"github.com/makerdao/oracle-suite/pkg/log"
	"github.com/makerdao/oracle-suite/pkg/spectre"
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

func (c *Config) Configure(d Dependencies) (*spectre.Spectre, error) {
	sig, err := c.Ethereum.ConfigureSigner()
	if err != nil {
		return nil, err
	}
	cli, err := c.Ethereum.ConfigureEthereumClient(sig)
	if err != nil {
		return nil, err
	}
	tra, err := c.Transport.Configure(transportConfig.Dependencies{
		Context: d.Context,
		Signer:  sig,
		Feeds:   c.Feeds.Addresses(),
		Logger:  d.Logger,
	})
	if err != nil {
		return nil, err
	}
	return c.Spectre.Configure(spectreConfig.Dependencies{
		Signer:         sig,
		Transport:      tra,
		EthereumClient: cli,
		Logger:         d.Logger,
	}), nil
}
