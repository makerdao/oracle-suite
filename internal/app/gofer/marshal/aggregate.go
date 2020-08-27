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

package marshal

import (
	"fmt"
	"strings"
	"time"

	"github.com/makerdao/gofer/pkg/model"
)

// priceAggregate is the CLI JSON output format representation for
// gofer/model/PriceAggregate used for marshaling
type priceAggregate struct {
	// Type is equivalent to PriceModelName, type in this context refers to
	// aggregate type.
	Type      string           `json:"type"`
	Pair      pair             `json:"pair"`
	Price     float64          `json:"price"`
	Volume    float64          `json:"volume,omitempty"`
	Timestamp *time.Time       `json:"timestamp,omitempty"`
	Prices    []priceAggregate `json:"prices,omitempty"`
	// ErrMsg is set to non-empty if the aggregation couldn't be validated
	ErrMsg string `json:"error,omitempty"`
}

// Pair is the CLI JSON output format representation for gofer/model/Pair
type pair struct {
	Base  string `json:"base"`
	Quote string `json:"quote"`
}

// newPriceAggregate returns a new instance of priceAggregate created from
// gofer/model/PriceAggregate and an optional error
func newPriceAggregate(agg *model.PriceAggregate, err error) priceAggregate {
	var prices []priceAggregate
	for _, pa := range agg.Prices {
		prices = append(prices, newPriceAggregate(pa, nil))
	}
	var errMsg string
	if err != nil {
		errMsg = err.Error()
	}
	var timestamp *time.Time
	if agg.Timestamp != 0 {
		t := time.Unix(agg.Timestamp, 0)
		timestamp = &t
	}
	return priceAggregate{
		Type:      agg.PriceModelName,
		Pair:      newPair(agg.Pair),
		Price:     agg.Price,
		Volume:    agg.Volume,
		Timestamp: timestamp,
		Prices:    prices,
		ErrMsg:    errMsg,
	}
}

// toString returns priceAggregate shorthand representation, for readability and
// debugging, e.g. of median for two BTC/USD prices
//
// BTC/USD$9100.0=median(
//   BTC/USD$9000.0=exchange[coinbase](),
//   BTC/USD$9200.0=exchange[binance]())
func (agg priceAggregate) toString(depth int) string {
	var str strings.Builder
	str.WriteString(agg.Pair.String())
	str.WriteString("$")
	str.WriteString(fmt.Sprintf("%f", agg.Price))
	if agg.ErrMsg != "" {
		str.WriteString("<error:")
		str.WriteString(agg.ErrMsg)
		str.WriteString(">")
	}
	str.WriteString("=")
	str.WriteString(agg.Type)
	str.WriteString("(")
	count := len(agg.Prices)
	if count > 0 {
		for i, pa := range agg.Prices {
			str.WriteString("\n")
			for d := 0; d < depth+1; d++ {
				str.WriteString("  ")
			}
			str.WriteString(pa.toString(depth + 1))
			if i < count-1 {
				str.WriteString(", ")
			}
		}
	}
	str.WriteString(")")
	return str.String()
}

// newPair returns a new instance of Pair given a gofer/model/Pair
func newPair(p *model.Pair) pair {
	return pair{
		Base:  p.Base,
		Quote: p.Quote,
	}
}

// String returns a string representation of Pair
func (p *pair) String() string {
	return strings.ToUpper(fmt.Sprintf("%s/%s", p.Base, p.Quote))
}
