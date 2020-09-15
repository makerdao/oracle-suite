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
const fxURL = "https://api.exchangeratesapi.io/latest?symbols=%s&base=%s"

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

func (f *Fx) Call(ppps []*model.PotentialPricePoint) []CallResult {
	// Group PPPs by asset pair base.
	bases := map[string][]*model.PotentialPricePoint{}
	for _, pp := range ppps {
		base := pp.Pair.Base
		bases[base] = append(bases[base], pp)
	}

	results := []CallResult{}
	for base, ppps := range bases {
		// Make one request per asset pair base.
		crs, err := f.callByBase(base, ppps)
		if err != nil {
			// If callByBase fails wholesale, create a CallResult per PPP with the same
			// error.
			crs = []CallResult{}
			for _, pp := range ppps {
				crs = append(crs, newCallResult(pp, nil, err))
			}
		}
		results = append(results, crs...)
	}

	return results
}

func (f *Fx) callOne(pp *model.PotentialPricePoint) (*model.PricePoint, error) {
	err := model.ValidatePotentialPricePoint(pp)
	if err != nil {
		return nil, err
	}

	crs, err := f.callByBase(pp.Pair.Base, []*model.PotentialPricePoint{pp})
	if err != nil {
		return nil, err
	}

	return crs[0].PricePoint, crs[0].Error
}

func (f *Fx) getURL(base string, quotes []*model.PotentialPricePoint) string {
	symbols := []string{}
	for _, pp := range quotes {
		symbols = append(symbols, f.renameSymbol(pp.Pair.Quote))
	}
	return fmt.Sprintf(fxURL, strings.Join(symbols, ","), f.renameSymbol(base))
}

func (f *Fx) callByBase(base string, ppps []*model.PotentialPricePoint) ([]CallResult, error) {
	req := &query.HTTPRequest{
		URL: f.getURL(base, ppps),
	}

	// Make query.
	res := f.Pool.Query(req)
	if res == nil {
		return nil, errEmptyExchangeResponse
	}
	if res.Error != nil {
		return nil, res.Error
	}
	// Parse JSON.
	var resp fxResponse
	err := json.Unmarshal(res.Body, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse FX response: %w", err)
	}
	if resp.Rates == nil {
		return nil, fmt.Errorf("failed to parse FX response: %+v", resp)
	}

	results := make([]CallResult, len(ppps))
	for i, pp := range ppps {
		if price, ok := resp.Rates[f.renameSymbol(pp.Pair.Quote)]; ok {
			// Build PricePoint from exchange response.
			results[i] = newCallResult(
				pp,
				&model.PricePoint{
					Exchange:  pp.Exchange,
					Pair:      pp.Pair,
					Price:     price,
					Timestamp: time.Now().Unix(),
				},
				nil,
			)
		} else {
			// Missing quote in exchange response.
			results[i] = newCallResult(
				pp,
				nil,
				fmt.Errorf("no price for %s quote exist in response %s", pp.Pair.Quote, res.Body),
			)
		}
	}
	return results, nil
}
