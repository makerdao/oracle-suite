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

	transportConfig "github.com/makerdao/oracle-suite/internal/config/transport"
	"github.com/makerdao/oracle-suite/pkg/transport"

	"github.com/makerdao/oracle-suite/pkg/log"
)

type Config struct {
	Transport transportConfig.Transport `json:"transport"`
}

type Dependencies struct {
	Context context.Context
	Logger  log.Logger
}

func (c *Config) Configure(d Dependencies) (transport.Transport, error) {
	tra, err := c.Transport.ConfigureP2PBoostrap(transportConfig.BootstrapDependencies{
		Context: d.Context,
		Logger:  d.Logger,
	})
	if err != nil {
		return nil, err
	}
	return tra, nil
}
