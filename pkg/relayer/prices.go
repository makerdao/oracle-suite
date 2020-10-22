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
	"fmt"
	"math/big"
	"math/rand"
	"sort"
	"time"

	"github.com/makerdao/gofer/pkg/oracle/median"
)

type Prices struct {
	assetPair string
	timeout   time.Duration
	prices    []*median.Price
}

func NewPrices(assetPair string, timeout time.Duration) *Prices {
	return &Prices{
		assetPair: assetPair,
		timeout:   timeout,
	}
}

func (p *Prices) Add(price *median.Price) error {
	if price.AssetPair != p.assetPair {
		return fmt.Errorf(
			"incompatible asset pair, %s given but %s expected",
			price.AssetPair,
			p.assetPair,
		)
	}

	p.prices = append(p.prices, price)
	return nil
}

func (p *Prices) Get() []*median.Price {
	return p.prices
}

func (p *Prices) Len() int64 {
	return int64(len(p.Get()))
}

func (p *Prices) ClearExpired() {
	var prices []*median.Price
	for _, price := range p.prices {
		if price.Age.Add(p.timeout).After(time.Now()) {
			prices = append(prices, price)
		}
	}

	p.prices = prices
}

func (p *Prices) Truncate(n int64) {
	if int64(len(p.prices)) <= n {
		return
	}

	rand.Shuffle(len(p.prices), func(i, j int) {
		p.prices[i], p.prices[j] = p.prices[j], p.prices[i]
	})

	p.prices = p.prices[0:n]
}

func (p *Prices) Median() *big.Int {
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

func (p *Prices) Clear() {
	p.prices = nil
}
