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
	"github.com/makerdao/gofer/model"
	"github.com/makerdao/gofer/query"
	"strings"
	"time"
)

// Fx URL
const fxURL = "https://api.exchangeratesapi.io/latest?base=%s"

type fxResponse struct {
	Rates map[string]float64 `json:"rates"`
}

// Fx exchange handler
type Fx struct{}

// Call implementation
func (b *Fx) Call(pool query.WorkerPool, pp *model.PotentialPricePoint) (*model.PricePoint, error) {
	if pool == nil {
		return nil, errNoPoolPassed
	}
	err := model.ValidatePotentialPricePoint(pp)
	if err != nil {
		return nil, err
	}

	pair := strings.ToUpper(pp.Pair.Base)
	req := &query.HTTPRequest{
		URL: fmt.Sprintf(fxURL, pair),
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
	var resp fxResponse
	err = json.Unmarshal(res.Body, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to pargse fx response: %s", err)
	}
	if resp.Rates == nil {
		return nil, fmt.Errorf("failed to parse FX response %+v", resp)
	}
	price, ok := resp.Rates[pp.Pair.Quote]
	if !ok {
		return nil, fmt.Errorf("no price for %s quote exist in response %s", pp.Pair.Quote, res.Body)
	}
	// building PricePoint
	return &model.PricePoint{
		Exchange:  pp.Exchange,
		Pair:      pp.Pair,
		Price:     price,
		Timestamp: time.Now().Unix(),
	}, nil
}
