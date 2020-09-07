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

	"github.com/makerdao/gofer/internal/query"
	"github.com/makerdao/gofer/pkg/model"
)

// Gemini URL
const geminiURL = "https://api.gemini.com/v1/pubticker/%s"

type geminiResponse struct {
	Price  string `json:"last"`
	Ask    string `json:"ask"`
	Bid    string `json:"bid"`
	Volume struct {
		Timestamp int64 `json:"timestamp"`
	}
}

// Gemini exchange handler
type Gemini struct {
	Pool query.WorkerPool
}

func (g *Gemini) localPairName(pair *model.Pair) string {
	return strings.ToLower(pair.Base + pair.Quote)
}

func (g *Gemini) getURL(pp *model.PricePoint) string {
	return fmt.Sprintf(geminiURL, g.localPairName(pp.Pair))
}

func (g *Gemini) Fetch(pps []*model.PricePoint) {
	for _, pp := range pps {
		g.fetchOne(pp)
	}
}

func (g *Gemini) fetchOne(pp *model.PricePoint) {
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
	var resp geminiResponse
	err = json.Unmarshal(res.Body, &resp)
	if err != nil {
		pp.Error = fmt.Errorf("failed to parse gemini response: %w", err)
		return
	}
	// Parsing price from string
	price, err := strconv.ParseFloat(resp.Price, 64)
	if err != nil {
		pp.Error = fmt.Errorf("failed to parse price from gemini exchange %s", res.Body)
		return
	}
	// Parsing ask from string
	ask, err := strconv.ParseFloat(resp.Ask, 64)
	if err != nil {
		pp.Error = fmt.Errorf("failed to parse ask from gemini exchange %s", res.Body)
		return
	}
	// Parsing bid from string
	bid, err := strconv.ParseFloat(resp.Bid, 64)
	if err != nil {
		pp.Error = fmt.Errorf("failed to parse bid from gemini exchange %s", res.Body)
		return
	}

	pp.Price = price
	pp.Ask = ask
	pp.Bid = bid
	pp.Timestamp = resp.Volume.Timestamp / 1000
}
