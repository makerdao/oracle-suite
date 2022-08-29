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

package rpc

import (
	"github.com/kRoqmoq/oracle-suite/internal/gofer/marshal"
	"github.com/kRoqmoq/oracle-suite/pkg/gofer"
	"github.com/kRoqmoq/oracle-suite/pkg/gofer/graph/feeder"
	"github.com/kRoqmoq/oracle-suite/pkg/log"
)

type Nothing = struct{}

type API struct {
	gofer gofer.Gofer
	log   log.Logger
}

type FeedArg struct {
	Pairs []gofer.Pair
}

type FeedResp struct {
	Warnings feeder.Warnings
}

type NodesArg struct {
	Format marshal.FormatType
	Pairs  []gofer.Pair
}

type NodesResp struct {
	Pairs map[gofer.Pair]*gofer.Model
}

type PricesArg struct {
	Pairs []gofer.Pair
}

type PricesResp struct {
	Prices map[gofer.Pair]*gofer.Price
}

type PairsResp struct {
	Pairs []gofer.Pair
}

func (n *API) Models(arg *NodesArg, resp *NodesResp) error {
	n.log.WithField("pairs", arg.Pairs).Info("Models")
	pairs, err := n.gofer.Models(arg.Pairs...)
	if err != nil {
		return err
	}
	resp.Pairs = pairs
	return nil
}

func (n *API) Prices(arg *PricesArg, resp *PricesResp) error {
	n.log.WithField("pairs", arg.Pairs).Info("Prices")
	prices, err := n.gofer.Prices(arg.Pairs...)
	if err != nil {
		return err
	}
	resp.Prices = prices
	return nil
}

func (n *API) Pairs(_ *Nothing, resp *PairsResp) error {
	n.log.Info("Prices")
	pairs, err := n.gofer.Pairs()
	if err != nil {
		return err
	}
	resp.Pairs = pairs
	return nil
}
