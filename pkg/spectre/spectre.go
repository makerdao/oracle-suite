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

package spectre

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/makerdao/oracle-suite/pkg/datastore"
	"github.com/makerdao/oracle-suite/pkg/ethereum"
	"github.com/makerdao/oracle-suite/pkg/log"
	"github.com/makerdao/oracle-suite/pkg/oracle"
)

const LoggerTag = "SPECTRE"

type errNotEnoughPricesForQuorum struct {
	AssetPair string
}

func (e errNotEnoughPricesForQuorum) Error() string {
	return fmt.Sprintf(
		"unable to update the Oracle for %s pair, there is not enough prices to achieve a quorum",
		e.AssetPair,
	)
}

type errUnknownAsset struct {
	AssetPair string
}

func (e errUnknownAsset) Error() string {
	return fmt.Sprintf("pair %s does not exists", e.AssetPair)
}

type errNoPrices struct {
	AssetPair string
}

func (e errNoPrices) Error() string {
	return fmt.Sprintf("there is no prices in the datastore for %s pair", e.AssetPair)
}

type Spectre struct {
	ctx    context.Context
	mu     sync.Mutex
	doneCh chan struct{}

	signer    ethereum.Signer
	datastore datastore.Datastore
	interval  time.Duration
	log       log.Logger
	pairs     map[string]*Pair
}

type Config struct {
	Context context.Context

	Signer ethereum.Signer
	// Datastore provides prices for Spectre.
	Datastore datastore.Datastore
	// Interval describes how often we should try to update Oracles.
	Interval time.Duration
	// Pairs is the list supported pairs by Spectre with their configuration.
	Pairs []*Pair
	// Logger is a current logger interface used by the Spectre. The Logger is
	// required to monitor asynchronous processes.
	Logger log.Logger
}

type Pair struct {
	// AssetPair is the name of asset pair, e.g. ETHUSD.
	AssetPair string
	// OracleSpread is the minimum spread between the Oracle price and new price
	// required to send update.
	OracleSpread float64
	// OracleExpiration is the minimum time difference between the Oracle time
	// and current time required to send an update.
	OracleExpiration time.Duration
	// PriceExpiration is the maximum amount of time before price received
	// from the feeder will be considered as expired.
	PriceExpiration time.Duration
	// Median is the instance of the oracle.Median which is the interface for
	// the Oracle contract.
	Median oracle.Median
}

func NewSpectre(cfg Config) (*Spectre, error) {
	if cfg.Context == nil {
		return nil, errors.New("context must not be nil")
	}

	r := &Spectre{
		ctx:       cfg.Context,
		doneCh:    make(chan struct{}),
		signer:    cfg.Signer,
		datastore: cfg.Datastore,
		interval:  cfg.Interval,
		pairs:     make(map[string]*Pair),
		log:       cfg.Logger.WithField("tag", LoggerTag),
	}
	for _, p := range cfg.Pairs {
		r.pairs[p.AssetPair] = p
	}
	return r, nil
}

func (r *Spectre) Start() error {
	r.log.Info("Starting")

	// Handle context cancellation:
	go func() {
		defer func() { r.doneCh <- struct{}{} }()
		defer r.log.Info("Stopped")
		<-r.ctx.Done()
	}()

	r.relayerLoop()

	return nil
}

func (r *Spectre) Wait() {
	<-r.doneCh
}

// relay tries to update an Oracle contract for given pair. It'll return
// transaction hash or nil if there is no need to update Oracle.
func (r *Spectre) relay(assetPair string) (*ethereum.Hash, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	pair, ok := r.pairs[assetPair]
	if !ok {
		return nil, errUnknownAsset{AssetPair: assetPair}
	}

	prices := newPrices(r.datastore.Prices().AssetPair(assetPair))
	if prices == nil || prices.len() == 0 {
		return nil, errNoPrices{AssetPair: assetPair}
	}

	oracleQuorum, err := pair.Median.Bar(r.ctx)
	if err != nil {
		return nil, err
	}
	oracleTime, err := pair.Median.Age(r.ctx)
	if err != nil {
		return nil, err
	}
	oraclePrice, err := pair.Median.Val(r.ctx)
	if err != nil {
		return nil, err
	}

	// Clear expired prices:
	prices.clearOlderThan(time.Now().Add(-1 * pair.PriceExpiration))
	prices.clearOlderThan(oracleTime)

	// Use only a minimum prices required to achieve a quorum:
	prices.truncate(oracleQuorum)

	spread := prices.spread(oraclePrice)
	isExpired := oracleTime.Add(pair.OracleExpiration).Before(time.Now())
	isStale := spread >= pair.OracleSpread

	// Print logs:
	r.log.
		WithFields(log.Fields{
			"assetPair":        assetPair,
			"bar":              oracleQuorum,
			"age":              oracleTime.String(),
			"val":              oraclePrice.String(),
			"expired":          isExpired,
			"stale":            isStale,
			"oracleExpiration": pair.OracleExpiration.String(),
			"oracleSpread":     pair.OracleSpread,
			"timeToExpiration": time.Since(oracleTime).String(),
			"currentSpread":    spread,
		}).
		Debug("Trying to update Oracle")
	for _, price := range prices.oraclePrices() {
		r.log.
			WithFields(price.Fields(r.signer)).
			Debug("Feed")
	}

	if isExpired || isStale {
		// Check if there are enough prices to achieve a quorum:
		if int64(prices.len()) != oracleQuorum {
			return nil, errNotEnoughPricesForQuorum{AssetPair: assetPair}
		}

		// Send *actual* transaction to the Ethereum network:
		tx, err := pair.Median.Poke(r.ctx, prices.oraclePrices(), true)
		return tx, err
	}

	// There is no need to update Oracle:
	return nil, nil
}

// relayerLoop creates a asynchronous loop which tries to send an update
// to an Oracle contract at a specified interval.
func (r *Spectre) relayerLoop() {
	if r.interval == 0 {
		return
	}

	ticker := time.NewTicker(r.interval)
	go func() {
		for {
			select {
			case <-r.doneCh:
				ticker.Stop()
				return
			case <-ticker.C:
				for assetPair := range r.pairs {
					tx, err := r.relay(assetPair)

					// Print log in case of an error:
					if err != nil {
						r.log.
							WithFields(log.Fields{"assetPair": assetPair}).
							WithError(err).
							Warn("Unable to update Oracle")
					}
					// Print log if there was no need to update prices:
					if err == nil && tx == nil {
						r.log.
							WithFields(log.Fields{"assetPair": assetPair}).
							Info("Oracle price is still valid")
					}
					// Print log if Oracle update transaction was sent:
					if tx != nil {
						r.log.
							WithFields(log.Fields{"assetPair": assetPair, "tx": tx.String()}).
							Info("Oracle updated")
					}
				}
			}
		}
	}()
}
