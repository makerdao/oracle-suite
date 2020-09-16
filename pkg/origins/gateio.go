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

// Gateio URL
const gateioSinglePairURL = "https://fx-api.gateio.ws/api/v4/spot/tickers?currency_pair=%s"
const gateioURL = "https://fx-api.gateio.ws/api/v4/spot/tickers"

// {"currency_pair":"LRC_USDT","last":"0.12176","lowest_ask":"0.12355","highest_bid":"0.12225",
//"change_percentage":"7.87",
//"base_volume":"2705363.321762761","quote_volume":"331862.539837944479403",
//"high_24h":"0.13315","low_24h":"0.10868"}
type gateioResponse struct {
	Pair   string `json:"currency_pair"`
	Volume string `json:"quote_volume"`
	Price  string `json:"last"`
	Ask    string `json:"lowest_ask"`
	Bid    string `json:"highest_bid"`
}

// Gateio exchange handler
type Gateio struct {
	Pool query.WorkerPool
}

func (g *Gateio) renameSymbol(symbol string) string {
	return strings.ToUpper(symbol)
}

func (g *Gateio) localPairName(pair *model.Pair) string {
	return fmt.Sprintf("%s_%s", g.renameSymbol(pair.Base), g.renameSymbol(pair.Quote))
}

func (g *Gateio) Call(ppps []*model.PotentialPricePoint) []CallResult {
	crs, err := g.call(ppps)
	if err != nil {
		return newCallResultErrors(ppps, err)
	}
	return crs
}

func (g *Gateio) call(ppps []*model.PotentialPricePoint) ([]CallResult, error) {
	var url string
	if len(ppps) == 1 {
		url = fmt.Sprintf(gateioSinglePairURL, g.localPairName(ppps[0].Pair))
	} else {
		url = gateioURL
	}

	req := &query.HTTPRequest{URL: url}

	// make query
	res := g.Pool.Query(req)
	if res == nil {
		return nil, errEmptyExchangeResponse
	}
	if res.Error != nil {
		return nil, res.Error
	}
	// parsing JSON
	var resps []gateioResponse
	err := json.Unmarshal(res.Body, &resps)
	if err != nil {
		return nil, fmt.Errorf("failed to parse gateio response: %w", err)
	}
	if len(resps) < 1 {
		return nil, fmt.Errorf("wrong gateio response: %s", res.Body)
	}

	respMap := map[string]gateioResponse{}
	for _, resp := range resps {
		respMap[resp.Pair] = resp
	}

	crs := make([]CallResult, len(ppps))
	for i, ppp := range ppps {
		symbol := g.localPairName(ppp.Pair)
		if resp, has := respMap[symbol]; has {
			p, err := g.newPricePoint(ppp, resp)
			if err != nil {
				crs[i] = newCallResultError(
					ppp,
					fmt.Errorf("failed to create price point from gateio response: %w: %s", err, res.Body),
				)
			} else {
				crs[i] = newCallResultSuccess(p)
			}
		} else {
			crs[i] = newCallResultError(
				ppp,
				fmt.Errorf("failed to find symbol %s in gateio response: %s", ppp.Pair, res.Body),
			)
		}
	}
	return crs, nil
}

func (g *Gateio) newPricePoint(pp *model.PotentialPricePoint, resp gateioResponse) (*model.PricePoint, error) {
	// Check pair name
	if resp.Pair != g.localPairName(pp.Pair) {
		return nil, fmt.Errorf("wrong gateio pair returned: %s", resp.Pair)
	}

	// Parsing price from string
	price, err := strconv.ParseFloat(resp.Price, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse price from gateio exchange")
	}
	// Parsing volume from string
	volume, err := strconv.ParseFloat(resp.Volume, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse volume from gateio exchange")
	}
	ask, err := strconv.ParseFloat(resp.Ask, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ask from gateio exchange")
	}
	bid, err := strconv.ParseFloat(resp.Bid, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse bid from gateio exchange")
	}

	// building PricePoint
	return &model.PricePoint{
		Exchange:  pp.Exchange,
		Pair:      pp.Pair,
		Price:     price,
		Volume:    volume,
		Ask:       ask,
		Bid:       bid,
		Timestamp: time.Now().Unix(),
	}, nil
}
