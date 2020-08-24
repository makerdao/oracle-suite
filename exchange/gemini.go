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

	"github.com/makerdao/gofer/model"
	"github.com/makerdao/gofer/query"
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

// LocalPairName implementation
func (g *Gemini) localPairName(pair *model.Pair) string {
	return strings.ToLower(pair.Base + pair.Quote)
}

// GetURL implementation
func (g *Gemini) getURL(pp *model.PotentialPricePoint) string {
	return fmt.Sprintf(geminiURL, g.localPairName(pp.Pair))
}

// Call implementation
func (g *Gemini) Call(pp *model.PotentialPricePoint) (*model.PricePoint, error) {
	err := model.ValidatePotentialPricePoint(pp)
	if err != nil {
		return nil, err
	}

	req := &query.HTTPRequest{
		URL: g.getURL(pp),
	}

	// make query
	res := g.Pool.Query(req)
	if res == nil {
		return nil, errEmptyExchangeResponse
	}
	if res.Error != nil {
		return nil, res.Error
	}
	// parsing JSON
	var resp geminiResponse
	err = json.Unmarshal(res.Body, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse gemini response: %w", err)
	}
	// Parsing price from string
	price, err := strconv.ParseFloat(resp.Price, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse price from gemini exchange %s", res.Body)
	}
	// Parsing ask from string
	ask, err := strconv.ParseFloat(resp.Ask, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ask from gemini exchange %s", res.Body)
	}
	// Parsing bid from string
	bid, err := strconv.ParseFloat(resp.Bid, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse bid from gemini exchange %s", res.Body)
	}
	// building PricePoint
	return &model.PricePoint{
		Exchange:  pp.Exchange,
		Pair:      pp.Pair,
		Price:     price,
		Ask:       ask,
		Bid:       bid,
		Timestamp: resp.Volume.Timestamp / 1000,
	}, nil
}
