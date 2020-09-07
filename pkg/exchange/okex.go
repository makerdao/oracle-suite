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
	"time"

	"github.com/makerdao/gofer/internal/query"
	"github.com/makerdao/gofer/pkg/model"
)

// Okex URL
const okexURL = "https://www.okex.com/api/spot/v3/instruments/%s/ticker"

type okexResponse struct {
	Last          string    `json:"last"`
	Ask           string    `json:"ask"`
	Bid           string    `json:"bid"`
	BaseVolume24H string    `json:"base_volume_24h"`
	Timestamp     time.Time `json:"timestamp"`
}

// Okex exchange handler
type Okex struct {
	Pool query.WorkerPool
}

func (o *Okex) getPair(pp *model.PricePoint) string {
	pair, ok := pp.Exchange.Config["pair"]
	if !ok || pair == "" {
		pair = o.localPairName(pp.Pair)
	}
	return pair
}

func (o *Okex) localPairName(pair *model.Pair) string {
	return fmt.Sprintf("%s-%s", pair.Base, pair.Quote)
}

func (o *Okex) getURL(pp *model.PricePoint) string {
	return fmt.Sprintf(okexURL, o.getPair(pp))
}

func (o *Okex) Fetch(pps []*model.PricePoint) {
	for _, pp := range pps {
		o.callOne(pp)
	}
}

func (o *Okex) callOne(pp *model.PricePoint) {
	err := model.ValidatePricePoint(pp)
	if err != nil {
		pp.Error = err
		return
	}

	req := &query.HTTPRequest{
		URL: o.getURL(pp),
	}

	// make query
	res := o.Pool.Query(req)
	if res == nil {
		pp.Error = errEmptyExchangeResponse
		return
	}
	if res.Error != nil {
		pp.Error = res.Error
		return
	}
	// parsing JSON
	var resp okexResponse
	err = json.Unmarshal(res.Body, &resp)
	if err != nil {
		pp.Error = fmt.Errorf("failed to parse okex response: %w", err)
		return
	}
	// parsing price from string
	price, err := strconv.ParseFloat(resp.Last, 64)
	if err != nil {
		pp.Error = fmt.Errorf("failed to parse price from okex exchange %s", res.Body)
		return
	}
	// parsing ask price from string
	ask, err := strconv.ParseFloat(resp.Ask, 64)
	if err != nil {
		pp.Error = fmt.Errorf("failed to parse ask from okex exchange %s", res.Body)
		return
	}
	// parsing bid price from string
	bid, err := strconv.ParseFloat(resp.Bid, 64)
	if err != nil {
		pp.Error = fmt.Errorf("failed to parse bid from okex exchange %s", res.Body)
		return
	}
	// parsing volume from string
	volume, err := strconv.ParseFloat(resp.BaseVolume24H, 64)
	if err != nil {
		pp.Error = fmt.Errorf("failed to parse volume from okex exchange %s", res.Body)
		return
	}

	pp.Timestamp = resp.Timestamp.Unix()

	pp.Price = price
	pp.Ask = ask
	pp.Bid = bid
	pp.Volume = volume
}
