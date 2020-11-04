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

// store contains a list of oracle.Price's for single asset pair. Only one
// price from single address can be added to that list.
type store struct {
	assetPair string
	prices    map[common.Address]*oracle.Price
}

// newStore creates the new Prices instance.
func newStore() *store {
	return &store{
		prices: make(map[common.Address]*oracle.Price, 0),
	}
}

// add adds a new price to the list. If a price from same address already
// exists, the newer one will be used.
func (p *store) add(price *oracle.Price) error {
	addr, err := price.From()
	if err != nil {
		return err
	}

	// Skip if previous price is newer:
	if prev, ok := p.prices[*addr]; ok && prev.Age.After(price.Age) {
		return nil
	}

	p.prices[*addr] = price
	return nil
}

// get returns all prices from the list.
func (p *store) get() []*oracle.Price {
	var prices []*oracle.Price
	for _, price := range p.prices {
		prices = append(prices, price)
	}

	return prices
}

// len returns the number of prices in the list.
func (p *store) len() int64 {
	return int64(len(p.prices))
}

// truncate removes random prices until the number of remaining prices is equal
// to n. If number of prices is less or equal to n, it does nothing.
//
// This method is used to reduce number of arguments in transaction which will
// reduce transaction costs.
func (p *store) truncate(n int64) {
	if int64(len(p.prices)) <= n {
		return
	}

	prices := p.get()
	rand.Shuffle(len(prices), func(i, j int) {
		prices[i], prices[j] = prices[j], prices[i]
	})

	p.clear()
	for _, price := range prices {
		_ = p.add(price)
	}
}

// median calculates median price for all prices in the list.
func (p *store) Median() *big.Int {
	prices := p.get()

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

// spread calculates spread between given price and a median price. The spread
// is returned as percentage points.
func (p *store) spread(price *big.Int) float64 {
	oldPriceF := new(big.Float).SetInt(price)
	newPriceF := new(big.Float).SetInt(p.Median())

	x := new(big.Float).Sub(newPriceF, oldPriceF)
	x = new(big.Float).Quo(x, oldPriceF)
	x = new(big.Float).Mul(x, big.NewFloat(100))
	xf, _ := x.Float64()

	return xf
}

// clearOlderThan deletes prices which are older than given time.
func (p *store) clearOlderThan(t time.Time) {
	for address, price := range p.prices {
		if price.Age.Before(t) {
			delete(p.prices, address)
		}
	}
}

// clear deletes all prices from the list.
func (p *store) clear() {
	p.prices = make(map[common.Address]*oracle.Price, 0)
}
