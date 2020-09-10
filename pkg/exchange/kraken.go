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

package exchange

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/makerdao/gofer/internal/query"
)

// Kraken URL
const krakenURL = "https://api.kraken.com/0/public/Ticker?pair=%s"

type krakenPairResponse struct {
	Price  []string `json:"c"`
	Volume []string `json:"v"`
}

type krakenResponse struct {
	Errors []string `json:"error"`
	Result map[string]*krakenPairResponse
}

// Kraken exchange handler
type Kraken struct {
	Pool query.WorkerPool
}

func (k *Kraken) getSymbol(symbol string) string {
	symbol = strings.ToUpper(symbol)

	// https://supairort.kraken.com/hc/en-us/articles/360001185506-How-to-interpret-asset-codes
	switch symbol {
	case "BTC":
		return "XXBT"
	case "DOGE":
		return "XXDG"
	default:
		prefixedSymbols := []string{
			"XETC",
			"XETH",
			"XLTC",
			"XMLN",
			"XREP",
			"XREPV2",
			"XXLM",
			"XXMR",
			"XXRP",
			"XXTZ",
			"XZEC",
			"ZCAD",
			"ZEUR",
			"ZGBP",
			"ZJPY",
			"ZUSD",
		}

		for _, s := range prefixedSymbols {
			if s == "X"+symbol || s == "Z"+symbol {
				return s
			}
		}

		return symbol
	}
}

func (k *Kraken) localPairName(pair Pair) string {
	return fmt.Sprintf("%s%s", k.getSymbol(pair.Base), k.getSymbol(pair.Quote))
}

func (k *Kraken) getURL(pair Pair) string {
	return fmt.Sprintf(krakenURL, k.localPairName(pair))
}

func (k *Kraken) Call(pairs []Pair) []CallResult {
	return callSinglePairExchange(k, pairs)
}

func (k *Kraken) callOne(pair Pair) (*Tick, error) {
	var err error
	req := &query.HTTPRequest{
		URL: k.getURL(pair),
	}

	// make query
	res := k.Pool.Query(req)
	if res == nil {
		return nil, errEmptyExchangeResponse
	}
	if res.Error != nil {
		return nil, res.Error
	}
	// parsing JSON
	var resp krakenResponse
	err = json.Unmarshal(res.Body, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse kraken response: %w", err)
	}
	if len(resp.Errors) > 0 {
		return nil, fmt.Errorf("kraken API error: %s", strings.Join(resp.Errors, " "))
	}
	result, ok := resp.Result[k.localPairName(pair)]
	if !ok || result == nil {
		return nil, fmt.Errorf("wrong kraken exchange response. No resulting data %+v", resp)
	}
	if len(result.Price) == 0 || len(result.Volume) == 0 {
		return nil, fmt.Errorf("wrong kraken exchange response. No resulting pair %s data %+v", pair, result)
	}
	// Parsing price from string
	price, err := strconv.ParseFloat(result.Price[0], 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse price from kraken exchange %s", res.Body)
	}
	// Parsing volume from string
	volume, err := strconv.ParseFloat(result.Volume[0], 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse volume from kraken exchange %s", res.Body)
	}
	// building Tick
	return &Tick{
		Pair:      pair,
		Price:     price,
		Volume24h: volume,
		Timestamp: time.Now(),
	}, nil
}
