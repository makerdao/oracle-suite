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
	"time"

	"github.com/makerdao/gofer/internal/ethereum"
	"github.com/makerdao/gofer/internal/ethereum/geth"
	"github.com/makerdao/gofer/internal/log"
	"github.com/makerdao/gofer/pkg/ghost"
	"github.com/makerdao/gofer/pkg/gofer"
	"github.com/makerdao/gofer/pkg/transport"
	"github.com/makerdao/gofer/pkg/transport/p2p"
	"github.com/makerdao/gofer/pkg/transport/p2p/ethkey"
)

type JSON struct {
	Ethereum JSONEthereum        `json:"ethereum"`
	P2P      JSONP2P             `json:"p2p"`
	Options  JSONOptions         `json:"options"`
	Feeds    []string            `json:"feeds"`
	Pairs    map[string]JSONPair `json:"pairs"`
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

type JSONOptions struct {
	Interval int `json:"interval"`
}

type JSONPair struct {
	MsgExpiration int `json:"msgExpiration"`
}

type JSONConfigErr struct {
	Err error
}

type Dependencies struct {
	Context context.Context
	Gofer   *gofer.Gofer
	Logger  log.Logger
}

type Instances struct {
	Signer    ethereum.Signer
	Transport transport.Transport
	Ghost     *ghost.Ghost
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

func (j *JSON) Configure(deps Dependencies) (*Instances, error) {
	// Create wallet for given account and keystore:
	acc, err := geth.NewAccount(
		j.Ethereum.Keystore,
		j.Ethereum.Password,
		ethereum.HexToAddress(j.Ethereum.From),
	)
	if err != nil {
		return nil, err
	}

	// Create new signer instance:
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

	// Create and configure Ghost:
	cfg := ghost.Config{
		Gofer:     deps.Gofer,
		Signer:    sig,
		Transport: tra,
		Logger:    deps.Logger,
		Interval:  time.Second * time.Duration(j.Options.Interval),
		Pairs:     nil,
	}
	for name, pair := range j.Pairs {
		cfg.Pairs = append(cfg.Pairs, &ghost.Pair{
			AssetPair:        name,
			OracleExpiration: time.Second * time.Duration(pair.MsgExpiration),
		})
	}
	gho, err := ghost.NewGhost(cfg)
	if err != nil {
		return nil, err
	}

	return &Instances{
		Signer:    sig,
		Transport: tra,
		Ghost:     gho,
	}, nil
}
