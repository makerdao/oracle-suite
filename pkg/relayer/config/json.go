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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/makerdao/gofer/internal/ethereum"
	"github.com/makerdao/gofer/internal/oracle"
	"github.com/makerdao/gofer/pkg/relayer"
	"github.com/makerdao/gofer/pkg/transport/serf"
)

type JSON struct {
	Ethereum JSONEthereum        `json:"ethereum"`
	Serf     JSONSerf            `json:"serf"`    // TODO
	Feeds    []string            `json:"feeds"`   // TODO
	Options  JSONOptions         `json:"options"` // TODO
	Pairs    map[string]JSONPair `json:"pairs"`
}

type JSONEthereum struct {
	From     string `json:"from"`
	Keystore string `json:"keystore"`
	Password string `json:"password"`
	RPC      string `json:"rpc"`
}

type JSONSerf struct {
	RPC string `json:"rpc"`
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

func (j *JSON) MakeRelayer() (*relayer.Relayer, error) {
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

	transport, err := serf.NewSerf(j.Serf.RPC, j.Options.MsgLimit)
	if err != nil {
		return nil, err
	}

	eth := ethereum.NewClient(client, wallet)
	rel := relayer.NewRelayer(transport, time.Second*time.Duration(j.Options.Interval))

	for name, pair := range j.Pairs {
		rel.AddPair(relayer.Pair{
			AssetPair:        name,
			OracleSpread:     pair.OracleSpread,
			OracleExpiration: time.Second * time.Duration(pair.OracleExpiration),
			PriceExpiration:  time.Second * time.Duration(pair.MsgExpiration),
			Median:           oracle.NewMedian(eth, common.HexToAddress(pair.Oracle), name),
		})
	}

	return rel, nil
}
