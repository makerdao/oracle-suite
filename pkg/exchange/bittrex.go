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
	"strings"
	"time"

	"github.com/makerdao/gofer/internal/query"
	"github.com/makerdao/gofer/pkg/model"
)

// BitTrex URL
const bittrexURL = "https://api.bittrex.com/api/v1.1/public/getticker?market=%s"

type bittrexResponse struct {
	Success bool `json:"success"`
	Result  struct {
		Ask  float64 `json:"Ask"`
		Bid  float64 `json:"Bid"`
		Last float64 `json:"Last"`
	} `json:"result"`
}

// BitTrex exchange handler
type BitTrex struct {
	Pool query.WorkerPool
}

func (b *BitTrex) localPairName(pair *model.Pair) string {
	return fmt.Sprintf("%s-%s", strings.ToUpper(pair.Quote), strings.ToUpper(pair.Base))
}

func (b *BitTrex) getURL(pp *model.PricePoint) string {
	return fmt.Sprintf(bittrexURL, b.localPairName(pp.Pair))
}

func (b *BitTrex) Fetch(ppps []*model.PricePoint) {
	for _, ppp := range ppps {
		b.callOne(ppp)
	}
}

func (b *BitTrex) callOne(pp *model.PricePoint) {
	err := model.ValidatePricePoint(pp)
	if err != nil {
		pp.Error = err
		return
	}

	req := &query.HTTPRequest{
		URL: b.getURL(pp),
	}

	// make query
	res := b.Pool.Query(req)
	if res == nil {
		pp.Error = errEmptyExchangeResponse
		return
	}
	if res.Error != nil {
		pp.Error = res.Error
		return
	}
	// parsing JSON
	var resp bittrexResponse
	err = json.Unmarshal(res.Body, &resp)
	if err != nil {
		pp.Error = fmt.Errorf("failed to parse bittrex response: %w", err)
		return
	}
	if !resp.Success {
		pp.Error = fmt.Errorf("wrong response from bittrex %v", resp)
		return
	}

	pp.Price = resp.Result.Last
	pp.Ask = resp.Result.Ask
	pp.Bid = resp.Result.Bid
	pp.Timestamp = time.Now().Unix()
}
