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

	"github.com/makerdao/gofer/pkg/ethereum"
	"github.com/makerdao/gofer/pkg/log"
	"github.com/makerdao/gofer/pkg/transport/messages"
)

const ClientLoggerTag = "SPIRE_CLIENT"

type Client struct {
	rpc     *rpc.Client
	network string
	address string
	signer  ethereum.Signer
	log     log.Logger
}

type ClientConfig struct {
	Signer           ethereum.Signer
	Network, address string
	Logger           log.Logger
}

func NewClient(cfg ClientConfig) *Client {
	return &Client{
		network: cfg.Network,
		address: cfg.address,
		signer:  cfg.Signer,
		log:     cfg.Logger.WithField("tag", ClientLoggerTag),
	}
}

func (s *Client) Start() error {
	s.log.Infof("Starting")

	client, err := rpc.DialHTTP(s.network, s.address)
	if err != nil {
		return err
	}
	s.rpc = client
	return nil
}

func (s *Client) Stop() error {
	defer s.log.Infof("Stopped")

	return s.rpc.Close()
}

func (s *Client) PublishPrice(price *messages.Price) error {
	s.log.
		WithFields(price.Price.Fields(s.signer)).
		Info("Publishing price")

	err := s.rpc.Call("API.PublishPrice", PublishPriceArg{Price: price}, &Nothing{})
	if err != nil {
		return err
	}
	return nil
}

func (s *Client) PullPrices(assetPair string) ([]*messages.Price, error) {
	s.log.
		WithField("assetPair", assetPair).
		Info("Pulling prices")

	resp := &PullPricesResp{}
	err := s.rpc.Call("API.PullPrices", PullPricesArg{AssetPair: assetPair}, resp)
	if err != nil {
		return nil, err
	}
	return resp.Prices, nil
}

func (s *Client) PullPrice(assetPair string, feeder string) (*messages.Price, error) {
	s.log.
		WithField("assetPair", assetPair).
		WithField("feeder", feeder).
		Info("Pulling price")

	resp := &PullPriceResp{}
	err := s.rpc.Call("API.PullPrice", PullPriceArg{AssetPair: assetPair, Feeder: feeder}, resp)
	if err != nil {
		return nil, err
	}
	return resp.Price, nil
}
