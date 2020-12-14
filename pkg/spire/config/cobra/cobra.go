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

package cobra

import (
	"github.com/spf13/pflag"

	"github.com/makerdao/gofer/pkg/spire/config"
)

func RegisterFlags(cfg *config.Config, flags *pflag.FlagSet) {
	flags.StringVar(&cfg.Ethereum.From, "config.ethereum.from", "", "")
	flags.StringVar(&cfg.Ethereum.Keystore, "config.ethereum.keystore", "", "")
	flags.StringVar(&cfg.Ethereum.Password, "config.ethereum.password", "", "")

	flags.StringSliceVar(&cfg.P2P.Listen, "config.p2p.listen", nil, "")
	flags.StringSliceVar(&cfg.P2P.BootstrapAddrs, "config.p2p.boostrap", nil, "")
	flags.StringSliceVar(&cfg.P2P.BlockedAddrs, "config.p2p.blocked", nil, "")

	flags.StringSliceVar(&cfg.Feeds, "config.feeds", nil, "")
	flags.StringSliceVar(&cfg.Feeds, "config.pairs", nil, "")
}
