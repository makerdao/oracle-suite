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
	"fmt"
	"strings"

	"github.com/makerdao/gofer/internal/query"
	"github.com/makerdao/gofer/pkg/model"
)

// Bitfinex URL
const bitfinexURL = "https://api-pub.bitfinex.com/v2/tickers/t%s"

// Bitfinex exchange handler
type Bitfinex struct {
	Pool query.WorkerPool
}

func (b *Bitfinex) localPairName(ppp *model.PotentialPricePoint) string {
	if ppp.Exchange != nil {
		if pn, ok := ppp.Exchange.Config["pair"]; ok {
			return pn
		}
	}

	const USDT = "USDT"
	const USD = "USD"
	if ppp.Pair.Base == USDT && ppp.Pair.Quote == USD {
		return "USTUSD"
	}
	if ppp.Pair.Quote == USDT {
		return ppp.Pair.Base + USD
	}

	return ppp.Pair.Base + ppp.Pair.Quote
}

func (b *Bitfinex) getURL(ppps []*model.PotentialPricePoint) string {
	var pairs []string
	for _, pp := range ppps {
		pairs = append(pairs, b.localPairName(pp))
	}

	return fmt.Sprintf(bitfinexURL, strings.Join(pairs, ","))
}

func (b *Bitfinex) Call(ppps []*model.PotentialPricePoint) ([]*model.PricePoint, error) {
	return nil, nil
	/*
	var err error

	// make query
	req := &query.HTTPRequest{URL: b.getURL(ppps)}
	res := b.Pool.Query(req)
	if res == nil {
		return nil, errEmptyExchangeResponse
	}
	if res.Error != nil {
		return nil, res.Error
	}

	// parsing JSON
	var resp [][]interface{}
	err = json.Unmarshal(res.Body, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse bitfinex response: %w", err)
	}
	if len(resp) < 8 {
		return nil, fmt.Errorf("wrong bitfinex response")
	}




	pps := make([]*model.PricePoint, 0)
	for _, ppp := range ppps {
		pp, err := b.call(ppp)
		if err != nil {
			return nil, err
		}

		pps = append(pps, pp)
	}

	return pps, nil
}

func (b *Bitfinex) call(pp *model.PotentialPricePoint) (*model.PricePoint, error) {
	err := model.ValidatePotentialPricePoint(pp)
	if err != nil {
		return nil, err
	}

	req := &query.HTTPRequest{
		URL: b.getURL(pp),
	}

	// make query
	res := b.Pool.Query(req)
	if res == nil {
		return nil, errEmptyExchangeResponse
	}
	if res.Error != nil {
		return nil, res.Error
	}

	// parsing JSON
	var resp [][]float64
	err = json.Unmarshal(res.Body, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse bitfinex response: %w", err)
	}
	if len(resp) < 8 {
		return nil, fmt.Errorf("wrong bitfinex response")
	}

	// building PricePoint
	return &model.PricePoint{
		Exchange:  pp.Exchange,
		Pair:      pp.Pair,
		Price:     resp[6],
		Volume:    resp[7],
		Timestamp: time.Now().Unix(),
	}, nil

	 */
}
