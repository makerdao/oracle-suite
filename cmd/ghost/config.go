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
	ghostConfig "github.com/makerdao/oracle-suite/internal/config/ghost"
	goferConfig "github.com/makerdao/oracle-suite/internal/config/gofer"
	transportConfig "github.com/makerdao/oracle-suite/internal/config/transport"
	"github.com/makerdao/oracle-suite/pkg/ghost"

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

func (c *Config) Configure(d Dependencies) (*ghost.Ghost, error) {
	gof, err := c.Gofer.ConfigureGofer(d.Logger, true)
	if err != nil {
		return nil, err
	}
	sig, err := c.Ethereum.ConfigureSigner()
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
	return c.Ghost.Configure(ghostConfig.Dependencies{
		Context:   d.Context,
		Gofer:     gof,
		Signer:    sig,
		Transport: tra,
		Logger:    d.Logger,
	})
}
