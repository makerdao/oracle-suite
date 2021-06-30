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

package spire

import (
	"net/rpc"

	"github.com/makerdao/oracle-suite/pkg/transport/messages"
)

type Spire struct {
	rpc     *rpc.Client
	network string
	address string
}

type Config struct {
	Network string
	Address string
}

func NewSpire(cfg Config) *Spire {
	return &Spire{
		network: cfg.Network,
		address: cfg.Address,
	}
}

func (s *Spire) Start() error {
	client, err := rpc.DialHTTP(s.network, s.address)
	if err != nil {
		return err
	}
	s.rpc = client
	return nil
}

func (s *Spire) Stop() error {
	return s.rpc.Close()
}

func (s *Spire) PublishPrice(price *messages.Price) error {
	err := s.rpc.Call("API.PublishPrice", PublishPriceArg{Price: price}, &Nothing{})
	if err != nil {
		return err
	}
	return nil
}

func (s *Spire) PullPrices(assetPair string, feeder string) ([]*messages.Price, error) {
	resp := &PullPricesResp{}
	err := s.rpc.Call("API.PullPrices", PullPricesArg{FilterAssetPair: assetPair, FilterFeeder: feeder}, resp)
	if err != nil {
		return nil, err
	}
	return resp.Prices, nil
}

func (s *Spire) PullPrice(assetPair string, feeder string) (*messages.Price, error) {
	resp := &PullPriceResp{}
	err := s.rpc.Call("API.PullPrice", PullPriceArg{AssetPair: assetPair, Feeder: feeder}, resp)
	if err != nil {
		return nil, err
	}
	return resp.Price, nil
}
