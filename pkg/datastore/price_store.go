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

// PriceStore contains a list of messages.Price's.
type PriceStore struct {
	mu sync.RWMutex

	prices map[string]map[ethereum.Address]*messages.Price
}

// NewPriceStore creates a new store instance.
func NewPriceStore() *PriceStore {
	return &PriceStore{
		prices: make(map[string]map[ethereum.Address]*messages.Price),
	}
}

// Add adds a new price to the list. If a price from same feeder already
// exists, the newer one will be used.
func (p *PriceStore) Add(from ethereum.Address, msg *messages.Price) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, ok := p.prices[msg.Price.Wat]; !ok {
		p.prices[msg.Price.Wat] = make(map[ethereum.Address]*messages.Price)
	}

	if prev, ok := p.prices[msg.Price.Wat][from]; ok && prev.Price.Age.After(msg.Price.Age) {
		return
	}

	p.prices[msg.Price.Wat][from] = msg
}

// AssetPair returns all prices for given asset pair.
func (p *PriceStore) AssetPair(assetPair string) *PriceSet {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if _, ok := p.prices[assetPair]; !ok {
		return NewPriceSet(nil)
	}

	var prices []*messages.Price
	for _, price := range p.prices[assetPair] {
		prices = append(prices, price)
	}

	return NewPriceSet(prices)
}

// Feeder returns the latest price for given asset pair sent by given feeder.
func (p *PriceStore) Feeder(assetPair string, feeder ethereum.Address) *messages.Price {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if _, ok := p.prices[assetPair]; !ok {
		return nil
	}
	if m, ok := p.prices[assetPair][feeder]; ok {
		return m
	}
	return nil
}
