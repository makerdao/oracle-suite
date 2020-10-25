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

	"github.com/makerdao/gofer/internal/ethereum"
	"github.com/makerdao/gofer/pkg/ghost"
	"github.com/makerdao/gofer/pkg/gofer"
	"github.com/makerdao/gofer/pkg/transport/serf"
)

type JSON struct {
	Ethereum JSONEthereum        `json:"ethereum"`
	Serf     JSONSerf            `json:"serf"`
	Options  JSONOptions         `json:"options"`
	Pairs    map[string]JSONPair `json:"pairs"`
}

type JSONEthereum struct {
	From     string `json:"from"`
	Keystore string `json:"keystore"`
	Password string `json:"password"`
}

type JSONSerf struct {
	RPC string `json:"rpc"`
}

type JSONOptions struct {
	Interval   int  `json:"interval"`
	SrcTimeout int  `json:"srcTimeout"` // TODO
	Verbose    bool `json:"verbose"`    // TODO
}

type JSONPair struct {
	MsgExpiration int     `json:"msgExpiration"`
	MsgSpread     float64 `json:"msgSpread"`
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

func (j *JSON) MakeGhost(gofer *gofer.Gofer) (*ghost.Ghost, error) {
	wallet, err := ethereum.NewWallet(
		j.Ethereum.Keystore,
		j.Ethereum.Password,
		common.HexToAddress(j.Ethereum.From),
	)
	if err != nil {
		return nil, err
	}

	transport, err := serf.NewSerf(j.Serf.RPC, 1024)
	if err != nil {
		return nil, err
	}

	gho := ghost.NewGhost(gofer, wallet, transport, time.Second*time.Duration(j.Options.Interval))
	for name, pair := range j.Pairs {
		err := gho.AddPair(ghost.Pair{
			AssetPair:        name,
			OracleSpread:     pair.MsgSpread,
			OracleExpiration: time.Second * time.Duration(pair.MsgExpiration),
		})
		if err != nil {
			return nil, err
		}
	}

	return gho, nil
}
