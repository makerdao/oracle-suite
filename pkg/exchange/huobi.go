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

func (h *Huobi) getURL(pp *model.PotentialPricePoint) string {
	return fmt.Sprintf(huobiURL, h.localPairName(pp.Pair))
}

func (h *Huobi) Call(ppps []*model.PotentialPricePoint) []CallResult {
	cr := make([]CallResult, 0)
	for _, ppp := range ppps {
		pp, err := h.callOne(ppp)

		cr = append(cr, CallResult{PricePoint: pp, Error: err})
	}

	return cr
}

func (h *Huobi) callOne(pp *model.PotentialPricePoint) (*model.PricePoint, error) {
	err := model.ValidatePotentialPricePoint(pp)
	if err != nil {
		return nil, err
	}

	req := &query.HTTPRequest{
		URL: h.getURL(pp),
	}

	res := h.Pool.Query(req)
	if res == nil {
		return nil, errEmptyExchangeResponse
	}
	if res.Error != nil {
		return nil, res.Error
	}

	var resp huobiResponse
	err = json.Unmarshal(res.Body, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse huobi response: %w", err)
	}
	if resp.Status == "error" {
		return nil, fmt.Errorf("wrong response from huobi exchange %s", res.Body)
	}
	if len(resp.Tick.Bid) < 1 {
		return nil, fmt.Errorf("wrong bid response from huobi exchange %s", res.Body)
	}

	return &model.PricePoint{
		Exchange:  pp.Exchange,
		Pair:      pp.Pair,
		Price:     resp.Tick.Bid[0],
		Volume:    resp.Volume,
		Timestamp: resp.Timestamp / 1000,
	}, nil
}
