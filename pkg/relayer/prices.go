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

package relayer

import (
	"math/big"
	"math/rand"
	"sort"
	"time"

	"github.com/makerdao/gofer/internal/oracle"
)

// prices contains a list of oracle.Price's for single asset pair.
type prices struct {
	assetPair string
	prices    []*oracle.Price
}

// newPrices creates the new Prices instance.
func newPrices() *prices {
	return &prices{
		prices: make([]*oracle.Price, 0),
	}
}

// Add adds a new price to the list.
func (p *prices) Add(price *oracle.Price) error {
	p.prices = append(p.prices, price)
	return nil
}

// Get returns all prices from the list.
func (p *prices) Get() []*oracle.Price {
	return p.prices
}

// Len returns the number of prices in the list.
func (p *prices) Len() int64 {
	return int64(len(p.Get()))
}

// Truncate removes random prices until the number of remaining prices is equal
// to n. If number of prices is less or equal to n, it does nothing.
//
// This method is used to reduce number of arguments in transaction which will
// reduce transaction costs.
func (p *prices) Truncate(n int64) {
	if int64(len(p.prices)) <= n {
		return
	}

	rand.Shuffle(len(p.prices), func(i, j int) {
		p.prices[i], p.prices[j] = p.prices[j], p.prices[i]
	})

	p.prices = p.prices[0:n]
}

// Median calculates median price for all prices in the list.
func (p *prices) Median() *big.Int {
	prices := p.Get()

	count := len(p.prices)
	if count == 0 {
		return big.NewInt(0)
	}

	sort.Slice(prices, func(i, j int) bool {
		return prices[i].Val.Cmp(prices[j].Val) < 0
	})

	if count%2 == 0 {
		m := count / 2
		x1 := prices[m-1].Val
		x2 := prices[m].Val
		return new(big.Int).Div(new(big.Int).Add(x1, x2), big.NewInt(2))
	}

	return prices[(count-1)/2].Val
}

// ClearOlderThan deletes expired prices form the list.
func (p *prices) ClearOlderThan(t time.Time) {
	var prices []*oracle.Price
	for _, price := range p.prices {
		if price.Age.After(t) {
			prices = append(prices, price)
		}
	}

	p.prices = prices
}

// Clear deletes all prices from the list.
func (p *prices) Clear() {
	p.prices = nil
}
