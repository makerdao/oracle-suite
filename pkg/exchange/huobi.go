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

	"github.com/makerdao/gofer/internal/query"
	"github.com/makerdao/gofer/pkg/model"
)

// Huobi URL
const huobiURL = "https://api.huobi.pro/market/detail/merged?symbol=%s"

type huobiResponse struct {
	Status    string  `json:"status"`
	Volume    float64 `json:"vol"`
	Timestamp int64   `json:"ts"`
	Tick      struct {
		Bid []float64
	}
}

// Huobi exchange handler
type Huobi struct {
	Pool query.WorkerPool
}

func (h *Huobi) localPairName(pair *model.Pair) string {
	return strings.ToLower(pair.Base + pair.Quote)
}

func (h *Huobi) getURL(pp *model.PricePoint) string {
	return fmt.Sprintf(huobiURL, h.localPairName(pp.Pair))
}

func (h *Huobi) Fetch(pps []*model.PricePoint) {
	for _, pp := range pps {
		h.fetchOne(pp)
	}
}

func (h *Huobi) fetchOne(pp *model.PricePoint) {
	err := model.ValidatePricePoint(pp)
	if err != nil {
		pp.Error = err
		return
	}

	req := &query.HTTPRequest{
		URL: h.getURL(pp),
	}

	res := h.Pool.Query(req)
	if res == nil {
		pp.Error = errEmptyExchangeResponse
		return
	}
	if res.Error != nil {
		pp.Error = res.Error
		return
	}

	var resp huobiResponse
	err = json.Unmarshal(res.Body, &resp)
	if err != nil {
		pp.Error = fmt.Errorf("failed to parse huobi response: %w", err)
		return
	}
	if resp.Status == "error" {
		pp.Error = fmt.Errorf("wrong response from huobi exchange %s", res.Body)
		return
	}
	if len(resp.Tick.Bid) < 1 {
		pp.Error = fmt.Errorf("wrong bid response from huobi exchange %s", res.Body)
		return
	}

	pp.Price = resp.Tick.Bid[0]
	pp.Volume = resp.Volume
	pp.Timestamp = resp.Timestamp / 1000
}
