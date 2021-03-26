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
	"net/rpc"

	"github.com/makerdao/oracle-suite/pkg/gofer"
)

// Gofer implements the gofer.Gofer interface. It uses a remote RPC server
// to fetch prices and models.
type Gofer struct {
	rpc     *rpc.Client
	network string
	address string
}

// Agent returns a new Gofer instance.
func NewGofer(network, address string) *Gofer {
	return &Gofer{
		network: network,
		address: address,
	}
}

// Start implements the gofer.StartableGofer interface.
func (c *Gofer) Start() error {
	client, err := rpc.DialHTTP(c.network, c.address)
	if err != nil {
		return err
	}
	c.rpc = client
	return nil
}

// Stop implements the gofer.StartableGofer interface.
func (c *Gofer) Stop() error {
	return c.rpc.Close()
}

// Models implements the gofer.Gofer interface.
func (c *Gofer) Models(pairs ...gofer.Pair) (map[gofer.Pair]*gofer.Model, error) {
	resp := &NodesResp{}
	err := c.rpc.Call("API.Models", NodesArg{Pairs: pairs}, resp)
	if err != nil {
		return nil, err
	}
	return resp.Pairs, nil
}

// Price implements the gofer.Gofer interface.
func (c *Gofer) Price(pair gofer.Pair) (*gofer.Price, error) {
	resp, err := c.Prices(pair)
	if err != nil {
		return nil, err
	}
	return resp[pair], nil
}

// Prices implements the gofer.Gofer interface.
func (c *Gofer) Prices(pairs ...gofer.Pair) (map[gofer.Pair]*gofer.Price, error) {
	resp := &PricesResp{}
	err := c.rpc.Call("API.Prices", PricesArg{Pairs: pairs}, resp)
	if err != nil {
		return nil, err
	}
	return resp.Prices, nil
}

// Pairs implements the gofer.Gofer interface.
func (c *Gofer) Pairs() ([]gofer.Pair, error) {
	resp := &PairsResp{}
	err := c.rpc.Call("API.Pairs", &Nothing{}, resp)
	if err != nil {
		return nil, err
	}
	return resp.Pairs, nil
}
