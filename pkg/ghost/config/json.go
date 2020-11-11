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

	"github.com/ethereum/go-ethereum/common"

	"github.com/makerdao/gofer/internal/ethereum"
	"github.com/makerdao/gofer/internal/log"
	"github.com/makerdao/gofer/internal/transport"
	"github.com/makerdao/gofer/internal/transport/p2p"
	"github.com/makerdao/gofer/pkg/ghost"
	"github.com/makerdao/gofer/pkg/gofer"
)

type JSON struct {
	Ethereum JSONEthereum        `json:"ethereum"`
	P2P      JSONP2P             `json:"p2p"`
	Options  JSONOptions         `json:"options"`
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
	MsgExpiration int     `json:"msgExpiration"`
	MsgSpread     float64 `json:"msgSpread"`
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
	Wallet    *ethereum.Wallet
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
	wal, err := ethereum.NewWallet(
		j.Ethereum.Keystore,
		j.Ethereum.Password,
		common.HexToAddress(j.Ethereum.From),
	)
	if err != nil {
		return nil, err
	}

	// Configure transport:
	tra, err := p2p.NewP2P(p2p.Config{
		Context:        deps.Context,
		ListenAddrs:    j.P2P.Listen,
		Wallet:         wal,
		BootstrapPeers: j.P2P.BootstrapPeers,
		BannedPeers:    j.P2P.BannedPeers,
		Logger:         deps.Logger,
	})
	if err != nil {
		return nil, err
	}

	// Create and configure Ghost:
	cfg := ghost.Config{
		Gofer:     deps.Gofer,
		Wallet:    wal,
		Transport: tra,
		Logger:    deps.Logger,
		Interval:  time.Second * time.Duration(j.Options.Interval),
		Pairs:     nil,
	}
	for name, pair := range j.Pairs {
		cfg.Pairs = append(cfg.Pairs, &ghost.Pair{
			AssetPair:        name,
			OracleSpread:     pair.MsgSpread,
			OracleExpiration: time.Second * time.Duration(pair.MsgExpiration),
		})
	}
	gho, err := ghost.NewGhost(cfg)
	if err != nil {
		return nil, err
	}

	return &Instances{
		Wallet:    wal,
		Transport: tra,
		Ghost:     gho,
	}, nil
}
