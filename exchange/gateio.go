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

// Gateio URL
const gateioURL = "https://fx-api.gateio.ws/api/v4/futures/tickers?contract=%s"

type gateioResponse struct {
	Volume string `json:"volume_24h_base"`
	Price  string `json:"last"`
}

// Gateio exchange handler
type Gateio struct{}

// GetURL implementation
func (g *Gateio) GetURL(pp *model.PotentialPricePoint) string {
	pair := fmt.Sprintf("%s_%s", strings.ToUpper(pp.Pair.Base), strings.ToUpper(pp.Pair.Quote))
	return fmt.Sprintf(gateioURL, pair)
}

// Call implementation
func (g *Gateio) Call(pool query.WorkerPool, pp *model.PotentialPricePoint) (*model.PricePoint, error) {
	if pool == nil {
		return nil, errNoPoolPassed
	}
	err := model.ValidatePotentialPricePoint(pp)
	if err != nil {
		return nil, err
	}

	req := &query.HTTPRequest{
		URL: g.GetURL(pp),
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
	var resp []gateioResponse
	err = json.Unmarshal(res.Body, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse gateio response: %w", err)
	}
	if len(resp) < 1 {
		return nil, fmt.Errorf("wrong gateio response %s", res.Body)
	}
	// Parsing price from string
	price, err := strconv.ParseFloat(resp[0].Price, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse price from gateio exchange %s", res.Body)
	}
	// Parsing price from string
	volume, err := strconv.ParseFloat(resp[0].Volume, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse volume from gateio exchange %s", res.Body)
	}
	// building PricePoint
	return &model.PricePoint{
		Exchange:  pp.Exchange,
		Pair:      pp.Pair,
		Price:     price,
		Volume:    volume,
		Timestamp: time.Now().Unix(),
	}, nil
}
