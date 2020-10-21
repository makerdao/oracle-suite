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
	"context"
	"errors"
	"fmt"
	"math/big"
	"math/rand"
	"sort"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/makerdao/gofer/pkg/oracle"
	"github.com/makerdao/gofer/pkg/oracle/median"
)

type Relayer struct {
	eth    *oracle.Ethereum
	wallet *oracle.Wallet
	median map[string]*median.Median
	pairs  map[string]Pair
	prices map[string][]*median.Price
}

type Pair struct {
	AssetPair        string
	Oracle           common.Address
	OracleSpread     float64
	OracleExpiration int64
	MsgExpiration    int64
}

func NewRelayer(eth *oracle.Ethereum, wallet *oracle.Wallet) *Relayer {
	return &Relayer{
		eth:    eth,
		wallet: wallet,
		median: make(map[string]*median.Median, 0),
		pairs:  make(map[string]Pair, 0),
		prices: make(map[string][]*median.Price, 0),
	}
}

func (r *Relayer) AddPair(pair Pair) {
	r.pairs[pair.AssetPair] = pair
	r.median[pair.AssetPair] = median.NewMedian(r.eth, pair.Oracle, pair.AssetPair)
}

func (r *Relayer) Pairs() map[string]Pair {
	return r.pairs
}

func (r *Relayer) Collect(price *median.Price) error {
	if price.Val.Cmp(big.NewInt(0)) == 0 {
		return errors.New("invalid price")
	}

	r.prices[price.AssetPair] = append(r.prices[price.AssetPair], price)

	return nil
}

func (r *Relayer) Relay(pair string) error {
	ctx := context.Background()

	m, ok := r.median[pair]
	if !ok {
		return fmt.Errorf("unable to find oracle contract for %s", pair)
	}

	// Check if the oracle price is expired:
	oracleTime, err := m.Age(ctx)
	if err != nil {
		return err
	}
	if time.Now().Unix()-oracleTime.Unix() < r.pairs[pair].OracleExpiration {
		return errors.New("unable to update oracle, price is not expired yet")
	}

	// Get non expired prices:
	var prices []*median.Price
	for _, p := range r.prices[pair] {
		if time.Now().Unix()-p.Time.Unix() < r.pairs[pair].MsgExpiration {
			prices = append(prices, p)
		}
	}

	// Check if there are enough prices to achieve a quorum:
	quorum, err := m.Bar(ctx)
	if err != nil {
		return err
	}
	if int64(len(prices)) < quorum {
		return errors.New("unable to update oracle, there is not enough prices to achieve a quorum")
	}

	// Use only a minimum prices required to achieve a quorum, this will save some gas:
	rand.Shuffle(len(prices), func(i, j int) { prices[i], prices[j] = prices[j], prices[i] })
	prices = prices[0:quorum]

	// Check if spread is large enough:
	medianPrice := calcMedian(prices)
	oldPrice, err := m.Price(ctx)
	if err != nil {
		return err
	}
	spread := calcSpread(oldPrice, medianPrice)
	if spread < r.pairs[pair].OracleSpread {
		return errors.New("unable to update oracle, spread is too low")
	}

	// Send transaction:
	_, err = m.Poke(ctx, prices)

	// Remove prices:
	r.prices[pair] = nil

	return err
}

func calcMedian(prices []*median.Price) *big.Int {
	count := len(prices)
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

func calcSpread(oldPrice, newPrice *big.Int) float64 {
	oldPriceF := new(big.Float).SetInt(oldPrice)
	newPriceF := new(big.Float).SetInt(newPrice)

	diff := new(big.Float).Sub(newPriceF, oldPriceF)
	div := new(big.Float).Quo(diff, oldPriceF)
	mul := new(big.Float).Mul(div, big.NewFloat(100))

	f, _ := mul.Float64()

	return f
}
