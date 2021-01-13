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
	"errors"
	"math/big"
	"sync"

	"github.com/makerdao/gofer/pkg/ethereum"
	"github.com/makerdao/gofer/pkg/log"
	"github.com/makerdao/gofer/pkg/transport"
	"github.com/makerdao/gofer/pkg/transport/messages"
)

const LoggerTag = "DATASTORE"

var errInvalidSignature = errors.New("received price has an invalid signature")
var errInvalidPrice = errors.New("received price is invalid")
var errUnknownPair = errors.New("received pair is not configured")
var errUnknownFeeder = errors.New("feeder is not allowed to send prices")

// Datastore reads and stores prices from the P2P network.
type Datastore struct {
	mu sync.Mutex

	signer     ethereum.Signer
	transport  transport.Transport
	pairs      map[string]*Pair
	priceStore *PriceStore
	log        log.Logger
	doneCh     chan struct{}
}

type Config struct {
	// Signer is an instance of the ethereum.Signer which will be used to
	// verify price signatures.
	Signer ethereum.Signer
	// Transport is a implementation of transport used to fetch prices from
	// feeders.
	Transport transport.Transport
	// Pairs is the list supported pairs by Spectre with their configuration.
	Pairs map[string]*Pair
	// Interval describes how often we should try to update Oracles.
	// Logger is a current logger interface used by the Datastore. The Logger is
	// required to monitor asynchronous processes.
	Logger log.Logger
}

type Pair struct {
	// Feeds is the list of Ethereum addresses from which prices will be
	// accepted.
	Feeds []ethereum.Address
}

func NewDatastore(config Config) *Datastore {
	return &Datastore{
		signer:     config.Signer,
		transport:  config.Transport,
		pairs:      config.Pairs,
		priceStore: NewPriceStore(),
		log:        config.Logger.WithField("tag", LoggerTag),
		doneCh:     make(chan struct{}),
	}
}

func (c *Datastore) Start() error {
	c.log.Info("Starting")

	c.mu.Lock()
	defer c.mu.Unlock()

	err := c.transport.Subscribe(messages.PriceMessageName)
	if err != nil {
		return err
	}

	return c.collectorLoop()
}

func (c *Datastore) Stop() error {
	defer c.log.Info("Stopped")

	close(c.doneCh)
	c.mu.Lock()
	defer c.mu.Unlock()

	err := c.transport.Unsubscribe(messages.PriceMessageName)
	if err != nil {
		return err
	}

	return nil
}

func (c *Datastore) Prices() *PriceStore {
	return c.priceStore
}

// collectPrice adds a price from a feeder which may be used to update
// Oracle contract. The price will be added only if a feeder is
// allowed to send prices.
func (c *Datastore) collectPrice(msg *messages.Price) error {
	from, err := msg.Price.From(c.signer)
	if err != nil {
		return errInvalidSignature
	}
	if _, ok := c.pairs[msg.Price.Wat]; !ok {
		return errUnknownPair
	}
	if !c.isFeedAllowed(msg.Price.Wat, *from) {
		return errUnknownFeeder
	}
	if msg.Price.Val.Cmp(big.NewInt(0)) <= 0 {
		return errInvalidPrice
	}

	c.priceStore.Add(*from, msg)

	return nil
}

// collectorLoop creates a asynchronous loop which fetches prices from feeders.
func (c *Datastore) collectorLoop() error {
	go func() {
		c.mu.Lock()
		defer c.mu.Unlock()

		for {
			price := &messages.Price{}
			select {
			case <-c.doneCh:
				return
			case status := <-c.transport.WaitFor(messages.PriceMessageName, price):
				// If there was a problem while reading prices from the transport:
				if status.Error != nil {
					c.log.
						WithError(status.Error).
						Warn("Unable to read prices from the transport")
					continue
				}

				// Try to collect received price:
				err := c.collectPrice(price)

				// Prepare log fields:
				from, _ := price.Price.From(c.signer)
				fields := log.Fields{"assetPair": price.Price.Wat}
				if from != nil {
					fields["from"] = from.String()
				}

				// Print logs:
				if err != nil {
					c.log.
						WithError(err).
						WithFields(fields).
						Warn("Received invalid price")
				} else {
					c.log.
						WithFields(fields).
						Info("Price received")
				}
			}
		}
	}()

	return nil
}

func (c *Datastore) isFeedAllowed(assetPair string, address ethereum.Address) bool {
	for _, a := range c.pairs[assetPair].Feeds {
		if a == address {
			return true
		}
	}
	return false
}
