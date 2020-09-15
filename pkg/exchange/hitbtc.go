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
const hitbtcURL = "https://api.hitbtc.com/api/2/public/ticker?symbols=%s"

type hitbtcResponse struct {
	Symbol    string    `json:"symbol"`
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

func (h *Hitbtc) getURL(ppps []*model.PotentialPricePoint) string {
	pairs := make([]string, len(ppps))
	for i, ppp := range ppps {
		pairs[i] = h.localPairName(ppp.Pair)
	}
	return fmt.Sprintf(hitbtcURL, strings.Join(pairs, ","))
}

func (h *Hitbtc) Call(ppps []*model.PotentialPricePoint) []CallResult {
	crs, err := h.call(ppps)
	if err != nil {
		return newCallResultErrors(ppps, err)
	}
	return crs

}

func (h *Hitbtc) call(ppps []*model.PotentialPricePoint) ([]CallResult, error) {
	req := &query.HTTPRequest{
		URL: h.getURL(ppps),
	}

	// make query
	res := h.Pool.Query(req)
	if res == nil {
		return nil, errEmptyExchangeResponse
	}
	if res.Error != nil {
		return nil, res.Error
	}
	// parsing JSON
	var resps []hitbtcResponse
	err := json.Unmarshal(res.Body, &resps)
	if err != nil {
		return nil, fmt.Errorf("failed to parse hitbtc response: %w", err)
	}

	respMap := map[string]hitbtcResponse{}
	for _, resp := range resps {
		respMap[resp.Symbol] = resp
	}

	crs := make([]CallResult, len(ppps))
	for i, ppp := range ppps {
		symbol := h.localPairName(ppp.Pair)
		if resp, has := respMap[symbol]; has {
			p, err := h.newPricePoint(ppp, resp)
			if err != nil {
				crs[i] = newCallResultError(
					ppp,
					fmt.Errorf("failed to create price point from hitbtc response: %w: %s", err, res.Body),
				)
			} else {
				crs[i] = newCallResultSuccess(p)
			}
		} else {
			crs[i] = newCallResultError(
				ppp,
				fmt.Errorf("failed to find symbol %s in hitbtc response: %s", ppp.Pair, res.Body),
			)
		}
	}
	return crs, nil
}

func (h *Hitbtc) newPricePoint(pp *model.PotentialPricePoint, resp hitbtcResponse) (*model.PricePoint, error) {
	// Parsing price from string.
	price, err := strconv.ParseFloat(resp.Price, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse price from hitbtc exchange")
	}
	// Parsing ask from string.
	ask, err := strconv.ParseFloat(resp.Ask, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ask from hitbtc exchange")
	}
	// Parsing volume from string.
	volume, err := strconv.ParseFloat(resp.Volume, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse volume from hitbtc exchange")
	}
	// Parsing bid from string.
	bid, err := strconv.ParseFloat(resp.Bid, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse bid from hitbtc exchange")
	}
	// Building PricePoint.
	return &model.PricePoint{
		Exchange:  pp.Exchange,
		Pair:      pp.Pair,
		Price:     price,
		Volume:    volume,
		Ask:       ask,
		Bid:       bid,
		Timestamp: resp.Timestamp.Unix(),
	}, nil
}
