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

// Hitbtc URL
const hitbtcURL = "https://api.hitbtc.com/api/2/public/ticker/%s"

type hitbtcResponse struct {
	Ask       string    `json:"ask"`
	Volume    string    `json:"volume"`
	Price     string    `json:"last"`
	Bid       string    `json:"bid"`
	Timestamp time.Time `json:"timestamp"`
}

// Hitbtc exchange handler
type Hitbtc struct {
	Pool query.WorkerPool
}

func (h *Hitbtc) localPairName(pair *model.Pair) string {
	return strings.ToUpper(pair.Base + pair.Quote)
}

func (h *Hitbtc) getURL(pp *model.PricePoint) string {
	return fmt.Sprintf(hitbtcURL, h.localPairName(pp.Pair))
}

func (h *Hitbtc) Fetch(pps []*model.PricePoint) {
	for _, pp := range pps {
		h.fetchOne(pp)
	}
}

//nolint:funlen
func (h *Hitbtc) fetchOne(pp *model.PricePoint) {
	err := model.ValidatePricePoint(pp)
	if err != nil {
		pp.Error = err
		return
	}

	req := &query.HTTPRequest{
		URL: h.getURL(pp),
	}

	// make query
	res := h.Pool.Query(req)
	if res == nil {
		pp.Error = errEmptyExchangeResponse
		return
	}
	if res.Error != nil {
		pp.Error = res.Error
		return
	}
	// parsing JSON
	var resp hitbtcResponse
	err = json.Unmarshal(res.Body, &resp)
	if err != nil {
		pp.Error = fmt.Errorf("failed to parse hitbtc response: %w", err)
		return
	}
	// Parsing price from string
	price, err := strconv.ParseFloat(resp.Price, 64)
	if err != nil {
		pp.Error = fmt.Errorf("failed to parse price from hitbtc exchange %s", res.Body)
		return
	}
	// Parsing ask from string
	ask, err := strconv.ParseFloat(resp.Ask, 64)
	if err != nil {
		pp.Error = fmt.Errorf("failed to parse ask from hitbtc exchange %s", res.Body)
		return
	}
	// Parsing volume from string
	volume, err := strconv.ParseFloat(resp.Volume, 64)
	if err != nil {
		pp.Error = fmt.Errorf("failed to parse volume from hitbtc exchange %s", res.Body)
		return
	}
	// Parsing bid from string
	bid, err := strconv.ParseFloat(resp.Bid, 64)
	if err != nil {
		pp.Error = fmt.Errorf("failed to parse bid from hitbtc exchange %s", res.Body)
		return
	}

	pp.Price = price
	pp.Volume = volume
	pp.Ask = ask
	pp.Bid = bid
	pp.Timestamp = resp.Timestamp.Unix()
}