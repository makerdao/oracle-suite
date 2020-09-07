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

// Poloniex URL
const poloniexURL = "https://poloniex.com/public?command=returnTicker"

type poloniexResponse struct {
	Price  string `json:"last"`
	Ask    string `json:"lowestAsk"`
	Bid    string `json:"highestBid"`
	Volume string `json:"baseVolume"`
}

// Poloniex exchange handler
type Poloniex struct {
	Pool query.WorkerPool
}

func (p *Poloniex) renameSymbol(symbol string) string {
	return strings.ToUpper(symbol)
}

func (p *Poloniex) localPairName(pair *model.Pair) string {
	return fmt.Sprintf("%s_%s", p.renameSymbol(pair.Quote), p.renameSymbol(pair.Base))
}

func (p *Poloniex) getURL(pp *model.PricePoint) string {
	return poloniexURL
}

func (p *Poloniex) Fetch(pps []*model.PricePoint) {
	for _, pp := range pps {
		p.fetchOne(pp)
	}
}

//nolint:funlen
func (p *Poloniex) fetchOne(pp *model.PricePoint) {
	err := model.ValidatePricePoint(pp)
	if err != nil {
		pp.Error = err
		return
	}

	req := &query.HTTPRequest{
		URL: p.getURL(pp),
	}

	pair := p.localPairName(pp.Pair)

	// make query
	res := p.Pool.Query(req)
	if res == nil {
		pp.Error = errEmptyExchangeResponse
		return
	}
	if res.Error != nil {
		pp.Error = res.Error
		return
	}
	// parsing JSON
	var resp map[string]poloniexResponse
	err = json.Unmarshal(res.Body, &resp)
	if err != nil {
		pp.Error = fmt.Errorf("failed to parse poloniex response: %w", err)
		return
	}
	pairResp, ok := resp[pair]
	if !ok {
		pp.Error = fmt.Errorf("failed to get correct response from exchange (no %s exist) %s", pair, res.Body)
		return
	}
	// Parsing price from string
	price, err := strconv.ParseFloat(pairResp.Price, 64)
	if err != nil {
		pp.Error = fmt.Errorf("failed to parse price from bitstamp exchange %s", res.Body)
		return
	}
	// Parsing ask from string
	ask, err := strconv.ParseFloat(pairResp.Ask, 64)
	if err != nil {
		pp.Error = fmt.Errorf("failed to parse ask from bitstamp exchange %s", res.Body)
		return
	}
	// Parsing volume from string
	volume, err := strconv.ParseFloat(pairResp.Volume, 64)
	if err != nil {
		pp.Error = fmt.Errorf("failed to parse volume from bitstamp exchange %s", res.Body)
		return
	}
	// Parsing bid from string
	bid, err := strconv.ParseFloat(pairResp.Bid, 64)
	if err != nil {
		pp.Error = fmt.Errorf("failed to parse bid from bitstamp exchange %s", res.Body)
		return
	}

	pp.Price = price
	pp.Volume = volume
	pp.Ask = ask
	pp.Bid = bid
	pp.Timestamp = time.Now().Unix()
}
