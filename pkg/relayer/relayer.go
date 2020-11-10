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
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/makerdao/gofer/internal/log"
	"github.com/makerdao/gofer/internal/oracle"
	"github.com/makerdao/gofer/internal/transport"
	"github.com/makerdao/gofer/pkg/messages"
)

const LoggerTag = "RELAYER"

type Relayer struct {
	mu        sync.Mutex
	feeds     []common.Address
	transport transport.Transport
	interval  time.Duration
	log       log.Logger
	pairs     map[string]Pair
	verbose   bool
	doneCh    chan struct{}
}

type Config struct {
	// Feeds is the list of Ethereum addresses from which prices will be
	// accepted.
	Feeds []string
	// Transport is a implementation of transport used to fetch prices from
	// feeders.
	Transport transport.Transport
	// Interval describes how often we should try to update Oracles.
	Interval time.Duration
	// Logger is a current logger interface used by the Relayer. The Logger is
	// required to monitor asynchronous processes.
	Logger log.Logger
	// Pairs is the list supported pairs by Relayer with their configuration.
	Pairs []Pair
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
	Median *oracle.Median
	// store contains list of prices form feeders.
	store *store
}

func NewRelayer(config Config) *Relayer {
	r := &Relayer{
		transport: config.Transport,
		interval:  config.Interval,
		pairs:     make(map[string]Pair, 0),
		log:       log.WrapLogger(config.Logger, log.Fields{"tag": LoggerTag}),
		doneCh:    make(chan struct{}),
	}

	for _, feed := range config.Feeds {
		r.feeds = append(r.feeds, common.HexToAddress(feed))
	}

	for _, pair := range config.Pairs {
		pair.store = newStore()
		r.pairs[pair.AssetPair] = pair
	}

	return r
}

func (r *Relayer) Start() error {
	r.log.Info("Starting")

	err := r.collectorLoop()
	if err != nil {
		return err
	}

	r.relayerLoop()
	return nil
}

func (r *Relayer) Stop() error {
	defer r.log.Info("Stopped")

	close(r.doneCh)
	err := r.transport.Unsubscribe(messages.PriceMessageName)
	if err != nil {
		return err
	}

	return nil
}

// collect adds a price from a feeder which may be used to update
// Oracle contract. The price will be added only if a feeder is
// allowed to send prices (must be on the r.Feeds list).
func (r *Relayer) collect(price *messages.Price) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	from, err := price.Price.From()
	if err != nil {
		return errors.New("received price has an invalid signature")
	}
	if !r.isFeedAllowed(*from) {
		return errors.New("address is not on feed list")
	}
	if price.Price.Val.Cmp(big.NewInt(0)) <= 0 {
		return errors.New("received price is invalid")
	}
	if _, ok := r.pairs[price.Price.AssetPair]; !ok {
		return errors.New("received pair is not configured")
	}

	err = r.pairs[price.Price.AssetPair].store.add(price.Price)
	if err != nil {
		return err
	}

	return nil
}

// relay tries to update an Oracle contract for given pair. It'll return
// transaction hash or nil if there is no need to update Oracle.
func (r *Relayer) relay(assetPair string) (*common.Hash, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	ctx := context.Background()
	pair := r.pairs[assetPair]

	oracleQuorum, err := pair.Median.Bar(ctx)
	if err != nil {
		return nil, err
	}
	oracleTime, err := pair.Median.Age(ctx)
	if err != nil {
		return nil, err
	}
	oraclePrice, err := pair.Median.Price(ctx)
	if err != nil {
		return nil, err
	}

	// Clear expired prices:
	pair.store.clearOlderThan(time.Now().Add(-1 * pair.PriceExpiration))
	pair.store.clearOlderThan(oracleTime)

	// Use only a minimum prices required to achieve a quorum:
	pair.store.truncate(oracleQuorum)

	spread := pair.store.spread(oraclePrice)
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
			"timeToExpiration": time.Now().Sub(oracleTime).String(),
			"currentSpread":    spread,
		}).
		Debug("Trying to update Oracle")
	for _, price := range pair.store.get() {
		r.log.
			WithFields(price.Fields()).
			Debug("Feed")
	}

	if isExpired || isStale {
		// Check if there are enough prices to achieve a quorum:
		if pair.store.len() != oracleQuorum {
			return nil, errors.New("unable to update Oracle, there is not enough prices to achieve a quorum")
		}

		// Send *actual* transaction to the Ethereum network:
		tx, err := pair.Median.Poke(ctx, pair.store.get(), true)
		// There is no point in keeping the prices that have already been sent,
		// so we can safely remove them:
		pair.store.clear()
		return tx, err
	}

	// There is no need to update Oracle:
	return nil, nil
}

// collectorLoop creates a asynchronous loop which fetches prices from feeders.
func (r *Relayer) collectorLoop() error {
	err := r.transport.Subscribe(messages.PriceMessageName)
	if err != nil {
		return err
	}

	go func() {
		for {
			price := &messages.Price{}
			select {
			case <-r.doneCh:
				return
			case status := <-r.transport.WaitFor(messages.PriceMessageName, price):
				// If there was a problem while reading prices from the transport:
				if status.Error != nil {
					r.log.
						WithError(status.Error).
						Warn("Unable to read prices from the network")
					continue
				}

				// Try to collect received price:
				err := r.collect(price)

				// Prepare log fields:
				from, _ := price.Price.From()
				fields := log.Fields{"assetPair": price.Price.AssetPair}
				if from != nil {
					fields["from"] = from.String()
				}

				// Print logs:
				if err != nil {
					r.log.
						WithError(err).
						WithFields(fields).
						Warn("Received invalid price")
				} else {
					r.log.
						WithFields(fields).
						Info("Price received")
				}
			}
		}
	}()

	return nil
}

// collectorLoop creates a asynchronous loop which tries to send an update
// to an Oracle contract at a specified interval.
func (r *Relayer) relayerLoop() {
	ticker := time.NewTicker(r.interval)
	go func() {
		for {
			select {
			case <-r.doneCh:
				ticker.Stop()
				return
			case <-ticker.C:
				for assetPair, _ := range r.pairs {
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

func (r *Relayer) isFeedAllowed(address common.Address) bool {
	for _, a := range r.feeds {
		if a == address {
			return true
		}
	}
	return false
}
