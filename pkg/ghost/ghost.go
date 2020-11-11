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

package ghost

import (
	"fmt"
	"sync"
	"time"

	"github.com/makerdao/gofer/internal/ethereum"
	"github.com/makerdao/gofer/internal/log"
	"github.com/makerdao/gofer/internal/marshal"
	"github.com/makerdao/gofer/internal/oracle"
	"github.com/makerdao/gofer/internal/transport"
	"github.com/makerdao/gofer/pkg/gofer"
	"github.com/makerdao/gofer/pkg/graph"
	"github.com/makerdao/gofer/pkg/messages"
)

const LoggerTag = "GHOST"

type Ghost struct {
	gofer     *gofer.Gofer
	wallet    *ethereum.Wallet
	transport transport.Transport
	interval  time.Duration
	pairs     map[graph.Pair]*Pair
	log       log.Logger
	doneCh    chan struct{}
}

type Config struct {
	// Gofer is an instance of the gofer.Gofer which will be used to fetch
	// prices.
	Gofer *gofer.Gofer
	// Wallet is an instance of the ethereum.Wallet which will be used to
	// sign prices.
	Wallet *ethereum.Wallet
	// Transport is a implementation of transport used to send prices to
	// relayers.
	Transport transport.Transport
	// Interval describes how often we should send prices to the network.
	Interval time.Duration
	// Logger is a current logger interface used by the Ghost. The Logger is
	//	// required to monitor asynchronous processes.
	Logger log.Logger
	// Pairs is the list supported pairs by Ghost with their configuration.
	Pairs []*Pair
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
}

func NewGhost(config Config) (*Ghost, error) {
	g := &Ghost{
		gofer:     config.Gofer,
		wallet:    config.Wallet,
		transport: config.Transport,
		interval:  config.Interval,
		pairs:     make(map[graph.Pair]*Pair, 0),
		log:       log.WrapLogger(config.Logger, log.Fields{"tag": LoggerTag}),
		doneCh:    make(chan struct{}),
	}

	// Unfortunately, the Gofer stores pairs in AAA/BBB format but Ghost (and
	// oracle contract) stores them in AAABBB format. Because of this we need
	// to make this wired mapping:
	for _, pair := range config.Pairs {
		found := false
		for _, goferPair := range g.gofer.Pairs() {
			if goferPair.Base+goferPair.Quote == pair.AssetPair {
				g.pairs[goferPair] = pair
				found = true
				break
			}
		}

		if !found {
			return nil, fmt.Errorf("unable to find the %s pair in the Gofer price models", pair.AssetPair)
		}
	}

	return g, nil
}

func (g *Ghost) Start() error {
	g.log.Infof("Starting")

	err := g.broadcasterLoop()
	if err != nil {
		return err
	}

	return nil
}

func (g *Ghost) Stop() error {
	defer g.log.Infof("Stopped")

	close(g.doneCh)
	err := g.transport.Unsubscribe(messages.PriceMessageName)
	if err != nil {
		return err
	}

	return nil
}

// broadcast sends price for single pair to the network. This method uses
// current price from the Gofer so it must be updated beforehand.
func (g *Ghost) broadcast(goferPair graph.Pair) error {
	var err error

	pair := g.pairs[goferPair]
	tick, err := g.gofer.Tick(goferPair)
	if err != nil {
		return err
	}
	if tick.Error != nil {
		return tick.Error
	}

	// Create price:
	price := oracle.NewPrice(pair.AssetPair)
	price.SetFloat64Price(tick.Price)
	price.Age = tick.Timestamp

	// Sign price:
	err = price.Sign(g.wallet)
	if err != nil {
		return err
	}

	// Broadcast price to P2P network:
	message, err := createPriceMessage(price, tick)
	if err != nil {
		return err
	}
	err = g.transport.Broadcast(messages.PriceMessageName, message)
	if err != nil {
		return err
	}

	return err
}

// broadcasterLoop creates a asynchronous loop which fetches prices from exchanges and then
// sends them to the network at a specified interval.
func (g *Ghost) broadcasterLoop() error {
	err := g.transport.Subscribe(messages.PriceMessageName)
	if err != nil {
		return err
	}

	ticker := time.NewTicker(g.interval)
	wg := sync.WaitGroup{}
	go func() {
		for {
			select {
			case <-g.doneCh:
				ticker.Stop()
				return
			case <-ticker.C:
				// Fetch prices from exchanges:
				err := g.gofer.Feed(g.gofer.Pairs()...)
				if err != nil {
					g.log.
						WithError(err).
						Warn("Unable to fetch prices for some pairs")
				}

				// Send prices to the network:
				//
				// Signing may be slow, especially with high KDF so this is why
				// we're using goroutines here.
				wg.Add(1)
				go func() {
					for assetPair, _ := range g.pairs {
						err := g.broadcast(assetPair)
						if err != nil {
							g.log.
								WithFields(log.Fields{"assetPair": assetPair}).
								WithError(err).
								Warn("Unable to broadcast price")
						} else {
							g.log.
								WithFields(log.Fields{"assetPair": assetPair}).
								Info("Price broadcasted")
						}
					}
					wg.Done()
				}()
			}
			wg.Wait()
		}
	}()

	return nil
}

func createPriceMessage(price *oracle.Price, tick graph.AggregatorTick) (*messages.Price, error) {
	trace, err := marshal.Marshall(marshal.JSON, tick)
	if err != nil {
		return nil, err
	}

	return &messages.Price{
		Price: price,
		Trace: trace,
	}, nil
}
