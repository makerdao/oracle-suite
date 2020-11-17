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

	"github.com/makerdao/gofer/internal/ethereum"
	"github.com/makerdao/gofer/internal/ethereum/geth"
	"github.com/makerdao/gofer/internal/log"
	"github.com/makerdao/gofer/internal/transport/p2p"
	"github.com/makerdao/gofer/internal/transport/p2p/ethkey"
	"github.com/makerdao/gofer/pkg/node"
)

type JSON struct {
	Ethereum JSONEthereum `json:"ethereum"`
	P2P      JSONP2P      `json:"p2p"`
	RPC      JSONRPC      `json:"rpc"`
	Feeds    []string     `json:"feeds"`
}

type JSONEthereum struct {
	From     string `json:"from"`
	Keystore string `json:"keystore"`
	Password string `json:"password"`
}

type JSONP2P struct {
	Listen         []string `json:"listen"`
	BootstrapPeers []string `json:"bootstrapPeers"`
	BannedPeers    []string `json:"bannedPeers"`
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

func (j *JSON) ConfigureServer(deps Dependencies) (*node.Server, error) {
	// Create wallet for a given account and keystore:
	acc, err := geth.NewAccount(
		j.Ethereum.Keystore,
		j.Ethereum.Password,
		ethereum.HexToAddress(j.Ethereum.From),
	)
	if err != nil {
		return nil, err
	}

	// Create a new signer instance:
	sig := geth.NewSigner(acc)

	// Configure transport:
	p2pCfg := p2p.Config{
		Context:        deps.Context,
		Signer:         sig,
		ListenAddrs:    j.P2P.Listen,
		BootstrapAddrs: j.P2P.BootstrapPeers,
		BlockedAddrs:   j.P2P.BannedPeers,
		Logger:         deps.Logger,
	}
	for _, feed := range j.Feeds {
		p2pCfg.AllowedPeers = append(p2pCfg.AllowedPeers, ethkey.AddressToPeerID(feed).Pretty())
	}
	tra, err := p2p.NewP2P(p2pCfg)
	if err != nil {
		return nil, err
	}

	// Create and configure RPC Server:
	srv, err := node.NewServer(tra, "tcp", j.RPC.Address)
	if err != nil {
		return nil, err
	}

	return srv, nil
}

func (j *JSON) ConfigureClient() (*node.Client, error) {
	return node.NewClient("tcp", j.RPC.Address), nil
}
