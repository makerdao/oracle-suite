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

package origins

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/makerdao/gofer/internal/query"
)

// Fx URL
const fxURL = "https://api.exchangeratesapi.io/latest?base=%s"

type fxResponse struct {
	Rates map[string]float64 `json:"rates"`
}

// Fx origin handler
type Fx struct {
	Pool query.WorkerPool
}

func (f *Fx) renameSymbol(symbol string) string {
	return strings.ToUpper(symbol)
}

func (f *Fx) localPairName(pair Pair) string {
	return f.renameSymbol(pair.Base)
}

func (f *Fx) getURL(pair Pair) string {
	return fmt.Sprintf(fxURL, f.localPairName(pair))
}

func (f *Fx) Fetch(pairs []Pair) []FetchResult {
	return callSinglePairOrigin(f, pairs)
}

func (f *Fx) callOne(pair Pair) (*Tick, error) {
	var err error
	req := &query.HTTPRequest{
		URL: f.getURL(pair),
	}

	// make query
	res := f.Pool.Query(req)
	if res == nil {
		return nil, errEmptyOriginResponse
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
	price, ok := resp.Rates[f.renameSymbol(pair.Quote)]
	if !ok {
		return nil, fmt.Errorf("no price for %s quote exist in response %s", pair.Quote, res.Body)
	}
	// building Tick
	return &Tick{
		Pair:      pair,
		Price:     price,
		Timestamp: time.Now(),
	}, nil
}
