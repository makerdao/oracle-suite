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
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/makerdao/gofer/internal/ethereum"
	"github.com/makerdao/gofer/internal/log"
	"github.com/makerdao/gofer/internal/oracle"
	"github.com/makerdao/gofer/internal/transport/p2p"
	"github.com/makerdao/gofer/pkg/relayer"
)

type JSON struct {
	Ethereum JSONEthereum        `json:"ethereum"`
	P2P      JSONP2P             `json:"p2p"`
	Feeds    []string            `json:"feeds"`
	Options  JSONOptions         `json:"options"` // TODO
	Pairs    map[string]JSONPair `json:"pairs"`
}

type JSONEthereum struct {
	From     string `json:"from"`
	Keystore string `json:"keystore"`
	Password string `json:"password"`
	RPC      string `json:"rpc"`
}

type JSONP2P struct {
	Listen         []string `json:"listen"`
	BootstrapPeers []string `json:"bootstrapPeers"`
	BannedPeers    []string `json:"bannedPeers"`
}

type JSONOptions struct {
	Interval int  `json:"interval"`
	MsgLimit int  `json:"msgLimit"`
	Verbose  bool `json:"verbose"` // TODO
}

type JSONPair struct {
	Oracle           string  `json:"oracle"`
	OracleSpread     float64 `json:"oracleSpread"`
	OracleExpiration int64   `json:"oracleExpiration"`
	MsgExpiration    int64   `json:"msgExpiration"`
}

type JSONConfigErr struct {
	Err error
}

type Dependencies struct {
	Context context.Context
	Logger  log.Logger
}

type Instances struct {
	P2P     *p2p.P2P
	Relayer *relayer.Relayer
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
	client, err := ethclient.Dial(j.Ethereum.RPC)
	if err != nil {
		return nil, err
	}

	wallet, err := ethereum.NewWallet(
		j.Ethereum.Keystore,
		j.Ethereum.Password,
		common.HexToAddress(j.Ethereum.From),
	)
	if err != nil {
		return nil, err
	}

	transport, err := p2p.NewP2P(p2p.Config{
		Context:        deps.Context,
		ListenAddrs:    j.P2P.Listen,
		BootstrapPeers: j.P2P.BootstrapPeers,
		BannedPeers:    j.P2P.BannedPeers,
		Logger:         deps.Logger,
	})
	if err != nil {
		return nil, err
	}

	eth := ethereum.NewClient(client, wallet)

	config := relayer.Config{
		Feeds:     j.Feeds,
		Transport: transport,
		Interval:  time.Second * time.Duration(j.Options.Interval),
		Logger:    deps.Logger,
		Pairs:     nil,
	}

	for name, pair := range j.Pairs {
		config.Pairs = append(config.Pairs, relayer.Pair{
			AssetPair:        name,
			OracleSpread:     pair.OracleSpread,
			OracleExpiration: time.Second * time.Duration(pair.OracleExpiration),
			PriceExpiration:  time.Second * time.Duration(pair.MsgExpiration),
			Median:           oracle.NewMedian(eth, common.HexToAddress(pair.Oracle), name),
		})
	}

	rel := relayer.NewRelayer(config)

	return &Instances{
		P2P:     transport,
		Relayer: rel,
	}, nil
}
