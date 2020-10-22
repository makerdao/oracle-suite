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
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/makerdao/gofer/pkg/oracle"
	"github.com/makerdao/gofer/pkg/oracle/median"
)

type Relayer struct {
	mu sync.Mutex

	eth    *oracle.Ethereum
	wallet *oracle.Wallet
	median map[string]*median.Median
	prices map[string]*Prices
	pairs  map[string]Pair
}

type Pair struct {
	AssetPair        string
	Oracle           common.Address
	OracleSpread     float64
	OracleExpiration time.Duration
	MsgExpiration    time.Duration
}

func NewRelayer(eth *oracle.Ethereum, wallet *oracle.Wallet) *Relayer {
	return &Relayer{
		eth:    eth,
		wallet: wallet,
		median: make(map[string]*median.Median, 0),
		prices: make(map[string]*Prices, 0),
		pairs:  make(map[string]Pair, 0),
	}
}

func (r *Relayer) AddPair(pair Pair) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.pairs[pair.AssetPair] = pair
	r.median[pair.AssetPair] = median.NewMedian(r.eth, pair.Oracle, pair.AssetPair)
}

func (r *Relayer) Pairs() map[string]Pair {
	return r.pairs
}

func (r *Relayer) Collect(price *median.Price) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if price.Val.Cmp(big.NewInt(0)) == 0 {
		return errors.New("invalid price")
	}

	if _, ok := r.prices[price.AssetPair]; !ok {
		r.prices[price.AssetPair] = NewPrices(price.AssetPair, r.pairs[price.AssetPair].MsgExpiration)
	}

	_ = r.prices[price.AssetPair].Add(price)

	return nil
}

func (r *Relayer) Relay(assetPair string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	ctx := context.Background()

	m, ok := r.median[assetPair]
	if !ok {
		return fmt.Errorf("unable to find oracle contract for %s", assetPair)
	}

	pair := r.pairs[assetPair]
	prices := r.prices[assetPair]
	prices.ClearExpired()

	// Check if the oracle price is expired:
	oracleTime, err := m.Age(ctx)
	if err != nil {
		return err
	}
	if oracleTime.Add(pair.OracleExpiration).After(time.Now()) {
		return errors.New("unable to update oracle, price is not expired yet")
	}

	// Check if there are enough prices to achieve a quorum:
	quorum, err := m.Bar(ctx)
	if err != nil {
		return err
	}
	if prices.Len() < quorum {
		return errors.New("unable to update oracle, there is not enough prices to achieve a quorum")
	}

	// Use only a minimum prices required to achieve a quorum, this will save some gas:
	prices.Truncate(quorum)

	// Check if spread is large enough:
	medianPrice := prices.Median()
	oldPrice, err := m.Price(ctx)
	if err != nil {
		return err
	}
	spread := calcSpread(oldPrice, medianPrice)
	if spread < pair.OracleSpread {
		return errors.New("unable to update oracle, spread is too low")
	}

	// Send transaction:
	_, err = m.Poke(ctx, prices.Get())

	// Remove prices:
	prices.Clear()

	return err
}

func calcSpread(oldPrice, newPrice *big.Int) float64 {
	oldPriceF := new(big.Float).SetInt(oldPrice)
	newPriceF := new(big.Float).SetInt(newPrice)

	x := new(big.Float).Sub(newPriceF, oldPriceF)
	x = new(big.Float).Quo(x, oldPriceF)
	x = new(big.Float).Mul(x, big.NewFloat(100))

	xf, _ := x.Float64()

	return xf
}
