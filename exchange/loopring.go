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

	"github.com/makerdao/gofer/model"
	"github.com/makerdao/gofer/query"
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
type Loopring struct{}

func (l *Loopring) localPairName(pair *model.Pair) string {
	return fmt.Sprintf("%s-%s", strings.ToUpper(pair.Base), strings.ToUpper(pair.Quote))
}

// Call implementation
func (l *Loopring) Call(pool query.WorkerPool, pp *model.PotentialPricePoint) (*model.PricePoint, error) {
	if pool == nil {
		return nil, errNoPoolPassed
	}
	err := model.ValidatePotentialPricePoint(pp)
	if err != nil {
		return nil, err
	}

	req := &query.HTTPRequest{
		URL: loopringURL,
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
