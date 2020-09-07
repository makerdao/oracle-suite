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
const gateioURL = "https://fx-api.gateio.ws/api/v4/spot/tickers?currency_pair=%s"

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

func (g *Gateio) getURL(pp *model.PricePoint) string {
	return fmt.Sprintf(gateioURL, g.localPairName(pp.Pair))
}

func (g *Gateio) Fetch(pps []*model.PricePoint) {
	for _, pp := range pps {
		g.callOne(pp)
	}
}

func (g *Gateio) callOne(pp *model.PricePoint) {
	err := model.ValidatePricePoint(pp)
	if err != nil {
		pp.Error = err
		return
	}

	req := &query.HTTPRequest{
		URL: g.getURL(pp),
	}

	// make query
	res := g.Pool.Query(req)
	if res == nil {
		pp.Error = errEmptyExchangeResponse
		return
	}
	if res.Error != nil {
		pp.Error = res.Error
		return
	}
	// parsing JSON
	var resp []gateioResponse
	err = json.Unmarshal(res.Body, &resp)
	if err != nil {
		pp.Error = fmt.Errorf("failed to parse gateio response: %w", err)
		return
	}
	if len(resp) < 1 {
		pp.Error = fmt.Errorf("wrong gateio response: %s", res.Body)
		return
	}
	// Check pair name
	if resp[0].Pair != g.localPairName(pp.Pair) {
		pp.Error = fmt.Errorf("wrong gateio pair returned %s: %s", resp[0].Pair, res.Body)
		return
	}

	// Parsing price from string
	price, err := strconv.ParseFloat(resp[0].Price, 64)
	if err != nil {
		pp.Error = fmt.Errorf("failed to parse price from gateio exchange: %s", res.Body)
		return
	}
	// Parsing volume from string
	volume, err := strconv.ParseFloat(resp[0].Volume, 64)
	if err != nil {
		pp.Error = fmt.Errorf("failed to parse volume from gateio exchange: %s", res.Body)
		return
	}
	ask, err := strconv.ParseFloat(resp[0].Ask, 64)
	if err != nil {
		pp.Error = fmt.Errorf("failed to parse ask from gateio exchange: %s", res.Body)
		return
	}
	bid, err := strconv.ParseFloat(resp[0].Bid, 64)
	if err != nil {
		pp.Error = fmt.Errorf("failed to parse bid from gateio exchange: %s", res.Body)
		return
	}

	pp.Price = price
	pp.Volume = volume
	pp.Ask = ask
	pp.Bid = bid
	pp.Timestamp = time.Now().Unix()
}
