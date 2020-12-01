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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/makerdao/gofer/pkg/datastore"
	"github.com/makerdao/gofer/pkg/ethereum"
	"github.com/makerdao/gofer/pkg/ethereum/geth"
	"github.com/makerdao/gofer/pkg/log"
	"github.com/makerdao/gofer/pkg/spire"
	"github.com/makerdao/gofer/pkg/transport"
	"github.com/makerdao/gofer/pkg/transport/p2p"
	"github.com/makerdao/gofer/pkg/transport/p2p/ethkey"
)

type JSON struct {
	Ethereum JSONEthereum `json:"ethereum"`
	P2P      JSONP2P      `json:"p2p"`
	RPC      JSONRPC      `json:"rpc"`
	Feeds    []string     `json:"feeds"`
	Pairs    []string     `json:"pairs"`
}

type JSONEthereum struct {
	From     string `json:"from"`
	Keystore string `json:"keystore"`
	Password string `json:"password"`
}

type JSONP2P struct {
	Listen         []string `json:"listen"`
	BootstrapAddrs []string `json:"bootstrapAddrs"`
	BlockedAddrs   []string `json:"blockedAddrs"`
}

type JSONRPC struct {
	Address string `json:"address"`
}

type JSONConfigErr struct {
	Err error
}

type Dependencies struct {
	Context context.Context
	Logger  log.Logger
}

func (e JSONConfigErr) Error() string {
	return e.Err.Error()
}

func ParseJSONFile(path string) (*JSON, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load JSON config file: %w", err)
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, JSONConfigErr{fmt.Errorf("failed to load JSON config file: %w", err)}
	}

	return ParseJSON(b)
}

func ParseJSON(b []byte) (*JSON, error) {
	j := &JSON{}
	err := json.Unmarshal(b, j)
	if err != nil {
		return nil, JSONConfigErr{err}
	}

	return j, nil
}

func (j *JSON) ConfigureServer(deps Dependencies) (*spire.Server, error) {
	// Ethereum account:
	acc, err := j.configureAccount()
	if err != nil {
		return nil, err
	}

	// Signer:
	sig := j.configureSigner(acc)

	// Transport:
	tra, err := j.configureTransport(deps.Context, sig, deps.Logger)
	if err != nil {
		return nil, err
	}

	// Datastore:
	dat := j.configureDatastore(sig, tra, deps.Logger)

	// RPC Server:
	srv, err := spire.NewServer(dat, tra, "tcp", j.RPC.Address)
	if err != nil {
		return nil, err
	}

	return srv, nil
}

func (j *JSON) ConfigureClient() (*spire.Client, error) {
	return spire.NewClient("tcp", j.RPC.Address), nil
}

func (j *JSON) configureAccount() (*geth.Account, error) {
	a, err := geth.NewAccount(
		j.Ethereum.Keystore,
		j.Ethereum.Password,
		ethereum.HexToAddress(j.Ethereum.From),
	)
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (j *JSON) configureSigner(a *geth.Account) ethereum.Signer {
	return geth.NewSigner(a)
}

func (j *JSON) configureTransport(ctx context.Context, s ethereum.Signer, l log.Logger) (transport.Transport, error) {
	cfg := p2p.Config{
		Context:        ctx,
		Signer:         s,
		ListenAddrs:    j.P2P.Listen,
		BootstrapAddrs: j.P2P.BootstrapAddrs,
		BlockedAddrs:   j.P2P.BlockedAddrs,
		Logger:         l,
	}
	for _, feed := range j.Feeds {
		cfg.AllowedPeers = append(cfg.AllowedPeers, ethkey.AddressToPeerID(feed).Pretty())
	}
	return p2p.New(cfg)
}

func (j *JSON) configureDatastore(s ethereum.Signer, t transport.Transport, l log.Logger) *datastore.Datastore {
	cfg := datastore.Config{
		Signer:    s,
		Transport: t,
		Pairs:     make(map[string]*datastore.Pair),
		Logger:    l,
	}
	var feeds []ethereum.Address
	for _, feed := range j.Feeds {
		feeds = append(feeds, ethereum.HexToAddress(feed))
	}
	for _, name := range j.Pairs {
		cfg.Pairs[name] = &datastore.Pair{Feeds: feeds}
	}
	return datastore.NewDatastore(cfg)
}
