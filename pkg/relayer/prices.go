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

	"github.com/ethereum/go-ethereum/common"

	"github.com/makerdao/gofer/internal/oracle"
)

// prices contains a list of oracle.Price's for single asset pair. Only one
// price from single address can be added to that list.
type prices struct {
	assetPair string
	prices    map[common.Address]*oracle.Price
}

// newPrices creates the new Prices instance.
func newPrices() *prices {
	return &prices{
		prices: make(map[common.Address]*oracle.Price, 0),
	}
}

// Add adds a new price to the list. If an price from same address already
// exists, it will be overwritten.
func (p *prices) Add(price *oracle.Price) error {
	addr, err := price.From()
	if err != nil {
		return err
	}

	p.prices[*addr] = price
	return nil
}

// Get returns all prices from the list.
func (p *prices) Get() []*oracle.Price {
	var prices []*oracle.Price
	for _, price := range p.prices {
		prices = append(prices, price)
	}

	return prices
}

// Len returns the number of prices in the list.
func (p *prices) Len() int64 {
	return int64(len(p.prices))
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

	prices := p.Get()
	rand.Shuffle(len(prices), func(i, j int) {
		prices[i], prices[j] = prices[j], prices[i]
	})

	p.Clear()
	for _, price := range prices {
		_ = p.Add(price)
	}
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

// Spread calculates spread between given price and an median. The spread is
// returned as percentage points.
func (p *prices) Spread(price *big.Int) float64 {
	oldPriceF := new(big.Float).SetInt(price)
	newPriceF := new(big.Float).SetInt(p.Median())

	x := new(big.Float).Sub(newPriceF, oldPriceF)
	x = new(big.Float).Quo(x, oldPriceF)
	x = new(big.Float).Mul(x, big.NewFloat(100))

	xf, _ := x.Float64()

	return xf
}

// ClearOlderThan deletes prices which are older than given time.
func (p *prices) ClearOlderThan(t time.Time) {
	for address, price := range p.prices {
		if price.Age.Before(t) {
			delete(p.prices, address)
		}
	}
}

// Clear deletes all prices from the list.
func (p *prices) Clear() {
	p.prices = make(map[common.Address]*oracle.Price, 0)
}
