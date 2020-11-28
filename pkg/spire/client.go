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

	"github.com/makerdao/gofer/pkg/transport/messages"
)

type Client struct {
	rpc     *rpc.Client
	network string
	address string
}

func NewClient(network, address string) *Client {
	return &Client{
		network: network,
		address: address,
	}
}

func (s *Client) Start() error {
	client, err := rpc.DialHTTP(s.network, s.address)
	if err != nil {
		return err
	}
	s.rpc = client
	return nil
}

func (s *Client) Stop() error {
	return s.rpc.Close()
}

func (s *Client) BroadcastPrice(price *messages.Price) error {
	err := s.rpc.Call("API.BroadcastPrice", price, &NoArgument{})
	if err != nil {
		return err
	}
	return nil
}

func (s *Client) GetPrices(assetPair string) ([]*messages.Price, error) {
	prices := &[]*messages.Price{}
	err := s.rpc.Call("API.GetPrices", assetPair, prices)
	if err != nil {
		return nil, err
	}
	return *prices, nil
}
