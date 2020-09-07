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

	"github.com/makerdao/gofer/internal/query"
	"github.com/makerdao/gofer/pkg/model"
)

// Bitstamp URL
const bitstampURL = "https://www.bitstamp.net/api/v2/ticker/%s"

type bitstampResponse struct {
	Ask       string `json:"ask"`
	Volume    string `json:"volume"`
	Price     string `json:"last"`
	Bid       string `json:"bid"`
	Timestamp string `json:"timestamp"`
}

// Bitstamp exchange handler
type Bitstamp struct {
	Pool query.WorkerPool
}

func (b *Bitstamp) renameSymbol(symbol string) string {
	return symbol
}

func (b *Bitstamp) localPairName(pair *model.Pair) string {
	return strings.ToLower(b.renameSymbol(pair.Base) + b.renameSymbol(pair.Quote))
}

func (b *Bitstamp) getURL(pp *model.PricePoint) string {
	return fmt.Sprintf(bitstampURL, b.localPairName(pp.Pair))
}

func (b *Bitstamp) Fetch(pps []*model.PricePoint) {
	for _, pp := range pps {
		b.fetchOne(pp)
	}
}

//nolint:funlen
func (b *Bitstamp) fetchOne(pp *model.PricePoint) {
	err := model.ValidatePricePoint(pp)
	if err != nil {
		pp.Error = err
		return
	}

	req := &query.HTTPRequest{
		URL: b.getURL(pp),
	}

	// make query
	res := b.Pool.Query(req)
	if res == nil {
		pp.Error = errEmptyExchangeResponse
		return
	}
	if res.Error != nil {
		pp.Error = res.Error
		return
	}
	// parsing JSON
	var resp bitstampResponse
	err = json.Unmarshal(res.Body, &resp)
	if err != nil {
		pp.Error = fmt.Errorf("failed to parse bitstamp response: %w", err)
		return
	}

	// Parsing price from string
	price, err := strconv.ParseFloat(resp.Price, 64)
	if err != nil {
		pp.Error = fmt.Errorf("failed to parse price from bitstamp exchange %s", res.Body)
		return
	}
	// Parsing ask from string
	ask, err := strconv.ParseFloat(resp.Ask, 64)
	if err != nil {
		pp.Error = fmt.Errorf("failed to parse ask from bitstamp exchange %s", res.Body)
		return
	}
	// Parsing volume from string
	volume, err := strconv.ParseFloat(resp.Volume, 64)
	if err != nil {
		pp.Error = fmt.Errorf("failed to parse volume from bitstamp exchange %s", res.Body)
		return
	}
	// Parsing bid from string
	bid, err := strconv.ParseFloat(resp.Bid, 64)
	if err != nil {
		pp.Error = fmt.Errorf("failed to parse bid from bitstamp exchange %s", res.Body)
		return
	}
	// Parsing timestamp from string
	timestamp, err := strconv.ParseInt(resp.Timestamp, 10, 64)
	if err != nil {
		pp.Error = fmt.Errorf("failed to parse timestamp from bitstamp exchange %s", res.Body)
		return
	}

	pp.Price = price
	pp.Volume = volume
	pp.Ask = ask
	pp.Bid = bid
	pp.Timestamp = timestamp
}
