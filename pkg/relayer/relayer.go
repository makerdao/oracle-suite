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
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/makerdao/gofer/internal/oracle"
	"github.com/makerdao/gofer/pkg/events"
	"github.com/makerdao/gofer/pkg/transport"
)

type Relayer struct {
	mu sync.Mutex

	transport transport.Transport
	interval  time.Duration
	feeds     []common.Address
	pairs     map[string]Pair
	doneCh    chan bool
}

type Pair struct {
	// AssetPair is the name of asset pair, e.g. ETHUSD.
	AssetPair string
	// OracleSpread is the minimum spread between the oracle price and new price
	// required to send update.
	OracleSpread float64
	// OracleExpiration is the minimum time difference between the oracle time
	// and current time required to send update.
	OracleExpiration time.Duration
	// PriceExpiration is the maximum TTL of the price from feeder.
	PriceExpiration time.Duration
	// Median is the instance of the oracle.Median which is the interface for
	// the median oracle contract.
	Median *oracle.Median
	// prices contains list of prices form the feeders.
	prices *prices
}

func NewRelayer(feeds []common.Address, transport transport.Transport, interval time.Duration) *Relayer {
	return &Relayer{
		feeds:     feeds,
		transport: transport,
		interval:  interval,
		pairs:     make(map[string]Pair, 0),
		doneCh:    make(chan bool),
	}
}

func (r *Relayer) AddPair(pair Pair) {
	r.mu.Lock()
	defer r.mu.Unlock()

	pair.prices = newPrices()
	r.pairs[pair.AssetPair] = pair
}

func (r *Relayer) Start(successCh chan<- string, errCh chan<- error) error {
	err := r.startCollector(errCh)
	if err != nil {
		return err
	}

	r.startRelayer(successCh, errCh)

	return nil
}

func (r *Relayer) Stop() {
	r.doneCh <- true
}

func (r *Relayer) collect(price *oracle.Price) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	from, err := price.From()
	if err != nil {
		return fmt.Errorf("recieved price has an invalid signature (pair: %s)", price.AssetPair)
	}
	if !onList(*from, r.feeds) {
		return fmt.Errorf("address is not on feeds list (pair: %s, from: %s)", price.AssetPair, from.String())
	}
	if price.Val.Cmp(big.NewInt(0)) <= 0 {
		return fmt.Errorf("recieved price is invalid (pair: %s, from: %s)", price.AssetPair, from.String())
	}
	if _, ok := r.pairs[price.AssetPair]; !ok {
		return fmt.Errorf("recieved pair is not configured (pair: %s, from: %s)", price.AssetPair, from.String())
	}

	err = r.pairs[price.AssetPair].prices.Add(price)
	if err != nil {
		return err
	}

	return nil
}

func (r *Relayer) relay(assetPair string) error {
	ctx := context.Background()
	pair := r.pairs[assetPair]

	oracleQuorum, err := pair.Median.Bar(ctx)
	if err != nil {
		return err
	}
	oracleTime, err := pair.Median.Age(ctx)
	if err != nil {
		return err
	}
	oraclePrice, err := pair.Median.Price(ctx)
	if err != nil {
		return err
	}

	// Clear expired prices:
	pair.prices.ClearOlderThan(time.Now().Add(-1 * pair.PriceExpiration))
	pair.prices.ClearOlderThan(oracleTime)

	// Use only a minimum prices required to achieve a quorum:
	pair.prices.Truncate(oracleQuorum)

	// Check if there are enough prices to achieve a quorum:
	if pair.prices.Len() != oracleQuorum {
		return fmt.Errorf(
			"unable to update the %s oracle, there is not enough prices to achieve a quorum (%d/%d)",
			assetPair,
			pair.prices.Len(),
			oracleQuorum,
		)
	}

	isExpired := oracleTime.Add(pair.OracleExpiration).After(time.Now())
	isStale := calcSpread(oraclePrice, pair.prices.Median()) >= pair.OracleSpread

	if isExpired || isStale {
		_, err = pair.Median.Poke(ctx, pair.prices.Get())
		pair.prices.Clear()
	}

	return fmt.Errorf("unable to update %s oracle: %w", assetPair, err)
}

func (r *Relayer) startCollector(onErrChan chan<- error) error {
	err := r.transport.Subscribe(events.PriceEventName)
	if err != nil {
		return err
	}

	go func() {
		for {
			price := &events.Price{}
			select {
			case <-r.doneCh:
				err := r.transport.Unsubscribe(events.PriceEventName)
				if err != nil {
					onErrChan <- err
				}

				return
			case status := <-r.transport.WaitFor(events.PriceEventName, price):
				if status.Error != nil && onErrChan != nil {
					onErrChan <- status.Error
					continue
				}
				err := r.collect(price.Price)
				if err != nil && onErrChan != nil {
					onErrChan <- err
				}
			}
		}
	}()

	return nil
}

func (r *Relayer) startRelayer(successCh chan<- string, errCh chan<- error) {
	ticker := time.NewTicker(r.interval)
	go func() {
		for {
			select {
			case <-r.doneCh:
				ticker.Stop()
				return
			case <-ticker.C:
				for assetPair, _ := range r.pairs {
					r.mu.Lock()
					err := r.relay(assetPair)
					if err != nil && errCh != nil {
						errCh <- err
					}
					if err == nil && successCh != nil {
						successCh <- assetPair
					}
					r.mu.Unlock()
				}
			}
		}
	}()
}

func onList(address common.Address, addresses []common.Address) bool {
	for _, a := range addresses {
		if a == address {
			return true
		}
	}
	return false
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
