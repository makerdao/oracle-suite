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
	"sync"

	"github.com/makerdao/gofer/pkg/ethereum"
	"github.com/makerdao/gofer/pkg/transport/messages"
)

type FeederPrice struct {
	AssetPair string
	Feeder    ethereum.Address
}

// PriceStore contains a list of messages.Price's.
type PriceStore struct {
	mu sync.RWMutex

	prices map[FeederPrice]*messages.Price
}

// NewPriceStore creates a new store instance.
func NewPriceStore() *PriceStore {
	return &PriceStore{
		prices: make(map[FeederPrice]*messages.Price),
	}
}

// Add adds a new price to the list. If a price from same feeder already
// exists, the newer one will be used.
func (p *PriceStore) Add(from ethereum.Address, msg *messages.Price) {
	p.mu.Lock()
	defer p.mu.Unlock()

	fp := FeederPrice{
		AssetPair: msg.Price.Wat,
		Feeder:    from,
	}

	if prev, ok := p.prices[fp]; ok && prev.Price.Age.After(msg.Price.Age) {
		return
	}

	p.prices[fp] = msg
}

// All returns all prices.
func (p *PriceStore) All() map[FeederPrice]*messages.Price {
	p.mu.Lock()
	defer p.mu.Unlock()

	r := map[FeederPrice]*messages.Price{}
	for k, v := range p.prices {
		r[k] = v
	}
	return r
}

// AssetPair returns all prices for given asset pair.
func (p *PriceStore) AssetPair(assetPair string) []*messages.Price {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var prices []*messages.Price
	for fp, price := range p.prices {
		if fp.AssetPair != assetPair {
			continue
		}
		prices = append(prices, price)
	}

	return prices
}

// Feeder returns the latest price for given asset pair sent by given feeder.
func (p *PriceStore) Feeder(assetPair string, feeder ethereum.Address) *messages.Price {
	p.mu.RLock()
	defer p.mu.RUnlock()

	fp := FeederPrice{
		AssetPair: assetPair,
		Feeder:    feeder,
	}

	if m, ok := p.prices[fp]; ok {
		return m
	}

	return nil
}
