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

package datastore

import (
	"math"
	"math/big"
	"math/rand"
	"sort"
	"time"

	"github.com/makerdao/gofer/pkg/oracle"
	"github.com/makerdao/gofer/pkg/transport/messages"
)

// PriceSet contains a list of messages.Price's for single asset pair. Only one
// price from a single address can be added to this list.
type PriceSet struct {
	msgs []*messages.Price
}

// NewPrices creates a new store instance.
func NewPriceSet(msgs []*messages.Price) *PriceSet {
	return &PriceSet{
		msgs: msgs,
	}
}

// Len returns the number of msgs in the list.
func (p *PriceSet) Len() int {
	return len(p.msgs)
}

// Messages returns raw messages list.
func (p *PriceSet) Messages() []*messages.Price {
	return p.msgs
}

// Prices returns oracle prices.
func (p *PriceSet) OraclePrices() []*oracle.Price {
	var prices []*oracle.Price
	for _, price := range p.msgs {
		prices = append(prices, price.Price)
	}
	return prices
}

// Truncate removes random msgs until the number of remaining msgs is equal
// to n. If number of msgs is less or equal to n, it does nothing.
//
// This method is used to reduce number of arguments in transaction which will
// reduce transaction costs.
func (p *PriceSet) Truncate(n int64) {
	if int64(len(p.msgs)) <= n {
		return
	}

	rand.Shuffle(len(p.msgs), func(i, j int) {
		p.msgs[i], p.msgs[j] = p.msgs[j], p.msgs[i]
	})

	p.msgs = p.msgs[0:n]
}

// Median calculates the Median price for all msgs in the list.
func (p *PriceSet) Median() *big.Int {
	count := len(p.msgs)
	if count == 0 {
		return big.NewInt(0)
	}

	sort.Slice(p.msgs, func(i, j int) bool {
		return p.msgs[i].Price.Val.Cmp(p.msgs[j].Price.Val) < 0
	})

	if count%2 == 0 {
		m := count / 2
		x1 := p.msgs[m-1].Price.Val
		x2 := p.msgs[m].Price.Val
		return new(big.Int).Div(new(big.Int).Add(x1, x2), big.NewInt(2))
	}

	return p.msgs[(count-1)/2].Price.Val
}

// Spread calculates the Spread between given price and a Median price.
// The Spread is returned as percentage points.
func (p *PriceSet) Spread(price *big.Int) float64 {
	if price.Cmp(big.NewInt(0)) == 0 {
		return math.Inf(1)
	}

	oldPriceF := new(big.Float).SetInt(price)
	newPriceF := new(big.Float).SetInt(p.Median())

	x := new(big.Float).Sub(newPriceF, oldPriceF)
	x = new(big.Float).Quo(x, oldPriceF)
	x = new(big.Float).Mul(x, big.NewFloat(100))
	xf, _ := x.Float64()

	return math.Abs(xf)
}

// ClearOlderThan deletes msgs which are older than given time.
func (p *PriceSet) ClearOlderThan(t time.Time) {
	var prices []*messages.Price
	for _, price := range p.msgs {
		if price.Price.Age.After(t) {
			prices = append(prices, price)
		}
	}
	p.msgs = prices
}

// Clear deletes all msgs from the list.
func (p *PriceSet) Clear() {
	p.msgs = nil
}
