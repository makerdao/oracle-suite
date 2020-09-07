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

	"github.com/makerdao/gofer/internal/query"
	"github.com/makerdao/gofer/pkg/model"
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

func (f *Fx) localPairName(pair *model.Pair) string {
	return f.renameSymbol(pair.Base)
}

func (f *Fx) getURL(pp *model.PricePoint) string {
	return fmt.Sprintf(fxURL, f.localPairName(pp.Pair))
}

func (f *Fx) Fetch(pps []*model.PricePoint) {
	for _, pp := range pps {
		f.fetchOne(pp)
	}
}

func (f *Fx) fetchOne(pp *model.PricePoint) {
	err := model.ValidatePricePoint(pp)
	if err != nil {
		pp.Error = err
		return
	}

	req := &query.HTTPRequest{
		URL: f.getURL(pp),
	}

	// make query
	res := f.Pool.Query(req)
	if res == nil {
		pp.Error = errEmptyExchangeResponse
		return
	}
	if res.Error != nil {
		pp.Error = res.Error
		return
	}
	// parsing JSON
	var resp fxResponse
	err = json.Unmarshal(res.Body, &resp)
	if err != nil {
		pp.Error = fmt.Errorf("failed to parse fx response: %w", err)
		return
	}
	if resp.Rates == nil {
		pp.Error = fmt.Errorf("failed to parse FX response %+v", resp)
		return
	}
	price, ok := resp.Rates[f.renameSymbol(pp.Pair.Quote)]
	if !ok {
		pp.Error = fmt.Errorf("no price for %s quote exist in response %s", pp.Pair.Quote, res.Body)
		return
	}
	// building PricePoint
	pp.Timestamp = time.Now().Unix()
	pp.Price = price
}
