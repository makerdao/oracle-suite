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
	"time"

	"github.com/makerdao/gofer/model"
	"github.com/makerdao/gofer/query"
)

// Exchange URL
const coinMarketCapURL = "https://pro-api.coinmarketcap.com/v1/cryptocurrency/quotes/latest?id=%s"

type quoteResponse struct {
	Price  float64 `json:"price"`
	Volume float64 `json:"volume_24h"`
}

type coinMarketCapPairResponse struct {
	Quote map[string]quoteResponse
}

type coinMarketCapResponse struct {
	Status struct {
		ErrorCode    int    `json:"error_code"`
		ErrorMessage string `json:"error_message"`
	}
	Data map[string]coinMarketCapPairResponse
}

// Exchange handler
type CoinMarketCap struct{}

// LocalPairName implementation
func (b *CoinMarketCap) LocalPairName(pair *model.Pair) string {
	switch *pair {
	case *model.NewPair("USDT", "USD"):
		return "825"
	case *model.NewPair("POLY", "USD"):
		return "2496"
	default:
		return pair.String()
	}
}

// GetURL implementation
func (b *CoinMarketCap) GetURL(pp *model.PotentialPricePoint) string {
	return fmt.Sprintf(coinMarketCapURL, b.LocalPairName(pp.Pair))
}

// Call implementation
func (b *CoinMarketCap) Call(pool query.WorkerPool, pp *model.PotentialPricePoint) (*model.PricePoint, error) {
	if pool == nil {
		return nil, errNoPoolPassed
	}
	err := model.ValidatePotentialPricePoint(pp)
	if err != nil {
		return nil, err
	}

	req := &query.HTTPRequest{
		URL: b.GetURL(pp),
	}

	// make query
	res := pool.Query(req)
	if res == nil {
		return nil, errEmptyExchangeResponse
	}
	if res.Error != nil {
		return nil, res.Error
	}
	// parsing JSON
	var resp coinMarketCapResponse
	err = json.Unmarshal(res.Body, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse CryptoCompare response: %w", err)
	}

	if resp.Status.ErrorCode != 0 || resp.Status.ErrorMessage != "" {
		return nil, fmt.Errorf("failed to get data from coinmarketcap: %s %s", resp.Status.ErrorMessage, res.Body)
	}
	id := b.LocalPairName(pp.Pair)
	pairResp, ok := resp.Data[id]

	if !ok {
		return nil, fmt.Errorf("failed to get quote pair %s, %s", pp.Pair.Quote, res.Body)
	}
	quoteRes, ok := pairResp.Quote[pp.Pair.Quote]
	if !ok {
		return nil, fmt.Errorf("failed to get quote response %s", res.Body)
	}

	// building PricePoint
	return &model.PricePoint{
		Exchange:  pp.Exchange,
		Pair:      pp.Pair,
		Price:     quoteRes.Price,
		Volume:    quoteRes.Volume,
		Timestamp: time.Now().Unix(),
	}, nil
}
