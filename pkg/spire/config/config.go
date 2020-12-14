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

package config

import (
	"context"

	"github.com/makerdao/gofer/pkg/datastore"
	"github.com/makerdao/gofer/pkg/ethereum"
	"github.com/makerdao/gofer/pkg/ethereum/geth"
	"github.com/makerdao/gofer/pkg/log"
	"github.com/makerdao/gofer/pkg/spire"
	"github.com/makerdao/gofer/pkg/transport"
	"github.com/makerdao/gofer/pkg/transport/p2p"
	"github.com/makerdao/gofer/pkg/transport/p2p/ethkey"
)

type Config struct {
	Ethereum Ethereum `json:"ethereum"`
	P2P      P2P      `json:"p2p"`
	RPC      RPC      `json:"rpc"`
	Feeds    []string `json:"feeds"`
	Pairs    []string `json:"pairs"`
}

type Ethereum struct {
	From     string `json:"from"`
	Keystore string `json:"keystore"`
	Password string `json:"password"`
}

type P2P struct {
	Listen         []string `json:"listen"`
	BootstrapAddrs []string `json:"bootstrapAddrs"`
	BlockedAddrs   []string `json:"blockedAddrs"`
}

type RPC struct {
	Address string `json:"address"`
}

type Dependencies struct {
	Context context.Context
	Logger  log.Logger
}

func (c *Config) ConfigureServer(deps Dependencies) (*spire.Server, error) {
	// Ethereum account:
	acc, err := c.configureAccount()
	if err != nil {
		return nil, err
	}

	// Signer:
	sig := c.configureSigner(acc)

	// Transport:
	tra, err := c.configureTransport(deps.Context, sig, deps.Logger)
	if err != nil {
		return nil, err
	}

	// Datastore:
	dat := c.configureDatastore(sig, tra, deps.Logger)

	// RPC Server:
	srv, err := spire.NewServer(spire.ServerConfig{
		Datastore: dat,
		Transport: tra,
		Signer:    sig,
		Network:   "tcp",
		Address:   c.RPC.Address,
		Logger:    deps.Logger,
	})
	if err != nil {
		return nil, err
	}

	return srv, nil
}

func (c *Config) ConfigureClient(deps Dependencies) (*spire.Client, error) {
	// Ethereum account:
	acc, err := c.configureAccount()
	if err != nil {
		return nil, err
	}

	// Signer:
	sig := c.configureSigner(acc)

	return spire.NewClient(spire.ClientConfig{
		Signer:  sig,
		Network: "tcp",
		Logger:  deps.Logger,
	}), nil
}

func (c *Config) configureAccount() (*geth.Account, error) {
	a, err := geth.NewAccount(
		c.Ethereum.Keystore,
		c.Ethereum.Password,
		ethereum.HexToAddress(c.Ethereum.From),
	)
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (c *Config) configureSigner(a *geth.Account) ethereum.Signer {
	return geth.NewSigner(a)
}

func (c *Config) configureTransport(ctx context.Context, s ethereum.Signer, l log.Logger) (transport.Transport, error) {
	cfg := p2p.Config{
		Context:        ctx,
		Signer:         s,
		ListenAddrs:    c.P2P.Listen,
		BootstrapAddrs: c.P2P.BootstrapAddrs,
		BlockedAddrs:   c.P2P.BlockedAddrs,
		Logger:         l,
	}
	for _, feed := range c.Feeds {
		cfg.AllowedPeers = append(cfg.AllowedPeers, ethkey.AddressToPeerID(feed).Pretty())
	}
	return p2p.New(cfg)
}

func (c *Config) configureDatastore(s ethereum.Signer, t transport.Transport, l log.Logger) *datastore.Datastore {
	cfg := datastore.Config{
		Signer:    s,
		Transport: t,
		Pairs:     make(map[string]*datastore.Pair),
		Logger:    l,
	}
	var feeds []ethereum.Address
	for _, feed := range c.Feeds {
		feeds = append(feeds, ethereum.HexToAddress(feed))
	}
	for _, name := range c.Pairs {
		cfg.Pairs[name] = &datastore.Pair{Feeds: feeds}
	}
	return datastore.NewDatastore(cfg)
}
