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
	"github.com/makerdao/gofer/internal/logger"
	"github.com/makerdao/gofer/internal/marshal"
	"github.com/makerdao/gofer/internal/oracle"
	"github.com/makerdao/gofer/internal/transport"
	"github.com/makerdao/gofer/pkg/events"
	"github.com/makerdao/gofer/pkg/gofer"
	"github.com/makerdao/gofer/pkg/graph"
)

type Ghost struct {
	gofer     *gofer.Gofer
	wallet    *ethereum.Wallet
	transport transport.Transport
	interval  time.Duration
	pairs     map[graph.Pair]Pair
	doneCh    chan bool
	logger    logger.Logger
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

func NewGhost(gofer *gofer.Gofer, wallet *ethereum.Wallet, transport transport.Transport, interval time.Duration, logger logger.Logger) *Ghost {
	return &Ghost{
		gofer:     gofer,
		wallet:    wallet,
		transport: transport,
		interval:  interval,
		pairs:     make(map[graph.Pair]Pair, 0),
		doneCh:    make(chan bool),
		logger:    logger,
	}
}

func (g *Ghost) AddPair(pair Pair) error {
	// Unfortunately, the Gofer stores pairs in AAA/BBB format but Ghost (and
	// oracle contract) stores them in AAABBB format. Because of this we need
	// to make this wired mapping:
	for _, goferPair := range g.gofer.Pairs() {
		if goferPair.Base+goferPair.Quote == pair.AssetPair {
			g.pairs[goferPair] = pair
			return nil
		}
	}

	return fmt.Errorf("unable to find the %s pair in the Gofer price models", pair.AssetPair)
}

func (g *Ghost) Start() error {
	err := g.transport.Subscribe(events.PriceEventName)
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
				err := g.gofer.Feed(g.gofer.Pairs()...)
				if err != nil {
					g.logger.Warning("GHOST", "Unable to fetch prices for some pairs: %s", err)
				}

				// Signing may be slow, especially with high KDF so this is why
				// we're using goroutines here:
				wg.Add(1)
				go func() {
					for assetPair, _ := range g.pairs {
						err := g.broadcast(assetPair)
						if err != nil {
							g.logger.Warning("GHOST", "Unable to broadcast price: %s", err)
						} else {
							g.logger.Info("GHOST", "Price broadcasted: %s", assetPair)
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

func (g *Ghost) Stop() error {
	err := g.transport.Unsubscribe(events.PriceEventName)
	if err != nil {
		return err
	}

	g.doneCh <- true
	return nil
}

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
	err = g.transport.Broadcast(events.PriceEventName, newPriceEvent(price, tick))
	if err != nil {
		return err
	}

	return err
}

func newPriceEvent(price *oracle.Price, tick graph.AggregatorTick) transport.Event {
	trace, _ := marshal.Marshall(marshal.JSON, tick)

	return &events.Price{
		Price: price,
		Trace: trace,
	}
}
