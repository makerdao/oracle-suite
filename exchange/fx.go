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
	"time"

	"github.com/makerdao/gofer/model"
	"github.com/makerdao/gofer/query"
)

// Fx URL
const fxURL = "https://api.exchangeratesapi.io/latest?base=%s"

type fxResponse struct {
	Rates map[string]float64 `json:"rates"`
}

// Fx exchange handler
type Fx struct {
	Pool query.WorkerPool
}

func (f *Fx) renameSymbol(symbol string) string {
	return strings.ToUpper(symbol)
}

// LocalPairName implementation
func (f *Fx) localPairName(pair *model.Pair) string {
	return f.renameSymbol(pair.Base)
}

// GetURL implementation
func (f *Fx) getURL(pp *model.PotentialPricePoint) string {
	return fmt.Sprintf(fxURL, f.localPairName(pp.Pair))
}

// Call implementation
func (f *Fx) Call(pp *model.PotentialPricePoint) (*model.PricePoint, error) {
	err := model.ValidatePotentialPricePoint(pp)
	if err != nil {
		return nil, err
	}

	req := &query.HTTPRequest{
		URL: f.getURL(pp),
	}

	// make query
	res := f.Pool.Query(req)
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
		return nil, fmt.Errorf("failed to parse fx response: %w", err)
	}
	if resp.Rates == nil {
		return nil, fmt.Errorf("failed to parse FX response %+v", resp)
	}
	price, ok := resp.Rates[f.renameSymbol(pp.Pair.Quote)]
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
