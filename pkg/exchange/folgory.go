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

// Folgory URL
const folgoryURL = "https://folgory.com/api/v1"

type folgoryResponse struct {
	Symbol string `json:"symbol"`
	Price  string `json:"last"`
	Volume string `json:"volume"`
}

// Folgory exchange handler
type Folgory struct {
	Pool query.WorkerPool
}

func (f *Folgory) renameSymbol(symbol string) string {
	return strings.ToUpper(symbol)
}

func (f *Folgory) localPairName(pair *model.Pair) string {
	return fmt.Sprintf("%s/%s", f.renameSymbol(pair.Base), f.renameSymbol(pair.Quote))
}

func (f *Folgory) Fetch(pps []*model.PricePoint) {
	for _, pp := range pps {
		f.fetchOne(pp)
	}
}

//nolint:funlen
func (f *Folgory) fetchOne(pp *model.PricePoint) {
	err := model.ValidatePricePoint(pp)
	if err != nil {
		pp.Error = err
		return
	}

	req := &query.HTTPRequest{
		URL: folgoryURL,
	}

	pair := f.localPairName(pp.Pair)

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
	var resp []folgoryResponse
	body := strings.TrimSpace(string(res.Body))

	err = json.Unmarshal([]byte(body), &resp)
	if err != nil {
		pp.Error = fmt.Errorf("failed to parse folgory response: %w", err)
		return
	}

	var data *folgoryResponse
	for _, symbol := range resp {
		if symbol.Symbol == pair {
			data = &symbol
			break
		}
	}
	if data == nil {
		pp.Error = fmt.Errorf("wrong response from folgory. no %s pair exist", pair)
		return
	}
	// Parsing price from string
	price, err := strconv.ParseFloat(data.Price, 64)
	if err != nil {
		pp.Error = fmt.Errorf("failed to parse price from folgory exchange %v", data)
		return
	}
	// Parsing volume from string
	volume, err := strconv.ParseFloat(data.Volume, 64)
	if err != nil {
		pp.Error = fmt.Errorf("failed to parse volume from folgory exchange %v", data)
		return
	}

	pp.Price = price
	pp.Volume = volume
	pp.Timestamp = time.Now().Unix()
}
