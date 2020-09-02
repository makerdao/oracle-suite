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

// Loopring URL
const loopringURL = "https://api.loopring.io/api/v2/overview"

type pairResponse struct {
	Price  string `json:"last"`
	Ask    string `json:"lowestAsk"`
	Bid    string `json:"highestBid"`
	Volume string `json:"quoteVolume"`
}

type loopringResponse struct {
	ResultInfo struct {
		Code    int
		Message string
	} `json:"resultInfo"`
	Data map[string]pairResponse `json:"data"`
}

// Loopring exchange handler
type Loopring struct {
	Pool query.WorkerPool
}

func (l *Loopring) localPairName(pair *model.Pair) string {
	return fmt.Sprintf("%s-%s", strings.ToUpper(pair.Base), strings.ToUpper(pair.Quote))
}

func (l *Loopring) getURL(pp *model.PotentialPricePoint) string {
	return loopringURL
}

func (l *Loopring) Call(ppps []*model.PotentialPricePoint) []CallResult {
	cr := make([]CallResult, 0)
	for _, ppp := range ppps {
		pp, err := l.call(ppp)

		cr = append(cr, CallResult{PricePoint: pp, Error: err})
	}

	return cr
}

func (l *Loopring) call(pp *model.PotentialPricePoint) (*model.PricePoint, error) {
	err := model.ValidatePotentialPricePoint(pp)
	if err != nil {
		return nil, err
	}
	req := &query.HTTPRequest{
		URL: l.getURL(pp),
	}
	// make query
	res := l.Pool.Query(req)
	if res == nil {
		return nil, errEmptyExchangeResponse
	}
	if res.Error != nil {
		return nil, res.Error
	}
	// parsing JSON
	var resp loopringResponse
	err = json.Unmarshal(res.Body, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse loopring response: %w", err)
	}
	if resp.ResultInfo.Code != 0 || resp.ResultInfo.Message != "SUCCESS" {
		return nil, fmt.Errorf("wrong loopring response %s", res.Body)
	}
	if resp.Data == nil {
		return nil, fmt.Errorf("empty `data` field for loopring response: %s", res.Body)
	}
	pair := l.localPairName(pp.Pair)
	pairRes, ok := resp.Data[pair]
	if !ok {
		return nil, fmt.Errorf("no %s pair exist in loopring response: %s", pair, res.Body)
	}
	// Parsing price from string
	price, err := strconv.ParseFloat(pairRes.Price, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse price from loopring exchange %s", res.Body)
	}
	// Parsing price from string
	volume, err := strconv.ParseFloat(pairRes.Volume, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse volume from loopring exchange %s", res.Body)
	}
	ask, err := strconv.ParseFloat(pairRes.Ask, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ask from loopring exchange %s", res.Body)
	}
	bid, err := strconv.ParseFloat(pairRes.Bid, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse bid from loopring exchange %s", res.Body)
	}
	// building PricePoint
	return &model.PricePoint{
		Exchange:  pp.Exchange,
		Pair:      pp.Pair,
		Price:     price,
		Volume:    volume,
		Ask:       ask,
		Bid:       bid,
		Timestamp: time.Now().Unix(),
	}, nil
}
