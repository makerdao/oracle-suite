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

	"github.com/makerdao/gofer/internal/query"
	"github.com/makerdao/gofer/pkg/model"
)

// Exchange URL
const cryptoCompareURL = "https://min-api.cryptocompare.com/data/price?fsym=%s&tsyms=%s"

type cryptoCompareResponse map[string]float64

// Exchange handler
type CryptoCompare struct {
	Pool query.WorkerPool
}

func (c *CryptoCompare) getURL(pp *model.PricePoint) string {
	return fmt.Sprintf(cryptoCompareURL, pp.Pair.Base, pp.Pair.Quote)
}

func (c *CryptoCompare) Fetch(pps []*model.PricePoint) {
	for _, pp := range pps {
		c.fetchOne(pp)
	}
}

func (c *CryptoCompare) fetchOne(pp *model.PricePoint) {
	err := model.ValidatePricePoint(pp)
	if err != nil {
		pp.Error = err
		return
	}

	req := &query.HTTPRequest{
		URL: c.getURL(pp),
	}

	// make query
	res := c.Pool.Query(req)
	if res == nil {
		pp.Error = errEmptyExchangeResponse
		return
	}
	if res.Error != nil {
		pp.Error = res.Error
		return
	}
	// parsing JSON
	var resp cryptoCompareResponse
	err = json.Unmarshal(res.Body, &resp)
	if err != nil {
		pp.Error = fmt.Errorf("failed to parse CryptoCompare response: %w", err)
		return
	}

	price, ok := resp[pp.Pair.Quote]
	if !ok {
		pp.Error = fmt.Errorf("failed to get price for %s: %s", pp.Pair.Quote, res.Body)
		return
	}

	// building PricePoint
	pp.Timestamp = time.Now().Unix()
	pp.Price = price
}
