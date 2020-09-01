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
	"github.com/makerdao/gofer/pkg/model"
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

func (k *Kraken) getPair(pp *model.PotentialPricePoint) string {
	if pp.Exchange == nil {
		return k.localPairName(pp.Pair)
	}

	pair, ok := pp.Exchange.Config["pair"]
	if !ok || pair == "" {
		pair = k.localPairName(pp.Pair)
	}
	return pair
}

func (k *Kraken) getSymbol(symbol string) string {
	symbol = strings.ToUpper(symbol)

	// https://support.kraken.com/hc/en-us/articles/360001185506-How-to-interpret-asset-codes
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

func (k *Kraken) localPairName(pair *model.Pair) string {
	return fmt.Sprintf("%s%s", k.getSymbol(pair.Base), k.getSymbol(pair.Quote))
}

func (k *Kraken) getURL(pp *model.PotentialPricePoint) string {
	return fmt.Sprintf(krakenURL, k.getPair(pp))
}

func (k *Kraken) Call(ppps []*model.PotentialPricePoint) ([]*model.PricePoint, error) {
	pps := make([]*model.PricePoint, 0)
	for _, ppp := range ppps {
		pp, err := k.call(ppp)
		if err != nil {
			return nil, err
		}

		pps = append(pps, pp)
	}

	return pps, nil
}

func (k *Kraken) call(pp *model.PotentialPricePoint) (*model.PricePoint, error) {
	err := model.ValidatePotentialPricePoint(pp)
	if err != nil {
		return nil, err
	}

	req := &query.HTTPRequest{
		URL: k.getURL(pp),
	}
	pair := k.getPair(pp)

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
	result, ok := resp.Result[pair]
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
	// building PricePoint
	return &model.PricePoint{
		Exchange:  pp.Exchange,
		Pair:      pp.Pair,
		Price:     price,
		Volume:    volume,
		Timestamp: time.Now().Unix(),
	}, nil
}
