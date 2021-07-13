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
	"context"
	"errors"
	"net/rpc"

	"github.com/makerdao/oracle-suite/pkg/gofer"
)

// Gofer implements the gofer.Gofer interface. It uses a remote RPC server
// to fetch prices and models.
type Gofer struct {
	ctx    context.Context
	doneCh chan struct{}

	rpc     *rpc.Client
	network string
	address string
}

// NewGofer returns a new Gofer instance.
func NewGofer(ctx context.Context, network, address string) (*Gofer, error) {
	if ctx == nil {
		return nil, errors.New("context must not be nil")
	}
	return &Gofer{
		ctx:     ctx,
		doneCh:  make(chan struct{}),
		network: network,
		address: address,
	}, nil
}

// Start implements the gofer.StartableGofer interface.
func (g *Gofer) Start() error {
	client, err := rpc.DialHTTP(g.network, g.address)
	if err != nil {
		return err
	}
	g.rpc = client

	go g.contextCancelHandler()
	return nil
}

// Wait implements the gofer.StartableGofer interface.
func (g *Gofer) Wait() {
	<-g.doneCh
}

// Models implements the gofer.Gofer interface.
func (g *Gofer) Models(pairs ...gofer.Pair) (map[gofer.Pair]*gofer.Model, error) {
	resp := &NodesResp{}
	err := g.rpc.Call("API.Models", NodesArg{Pairs: pairs}, resp)
	if err != nil {
		return nil, err
	}
	return resp.Pairs, nil
}

// Price implements the gofer.Gofer interface.
func (g *Gofer) Price(pair gofer.Pair) (*gofer.Price, error) {
	resp, err := g.Prices(pair)
	if err != nil {
		return nil, err
	}
	return resp[pair], nil
}

// Prices implements the gofer.Gofer interface.
func (g *Gofer) Prices(pairs ...gofer.Pair) (map[gofer.Pair]*gofer.Price, error) {
	resp := &PricesResp{}
	err := g.rpc.Call("API.Prices", PricesArg{Pairs: pairs}, resp)
	if err != nil {
		return nil, err
	}
	return resp.Prices, nil
}

// Pairs implements the gofer.Gofer interface.
func (g *Gofer) Pairs() ([]gofer.Pair, error) {
	resp := &PairsResp{}
	err := g.rpc.Call("API.Pairs", &Nothing{}, resp)
	if err != nil {
		return nil, err
	}
	return resp.Pairs, nil
}

func (g *Gofer) contextCancelHandler() {
	defer func() { g.doneCh <- struct{}{} }()
	<-g.ctx.Done()

	g.rpc.Close()
}
