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

package model

import (
	"fmt"
	"math"
)

const priceMultiplier = 10000000

// Pair represents currency pair
type Pair struct {
	Base  string
	Quote string
}

// Equal check if `Pair` is equal to given one
func (p *Pair) Equal(pair *Pair) bool {
	return (p.Base == pair.Base && p.Quote == pair.Quote)
}

// Exchange represent exchange details
type Exchange struct {
	Name   string
	Config map[string]string
}

// PricePoint given price point
type PricePoint struct {
	Timestamp int64     // Unix time
	Exchange  *Exchange // Exchange id
	Pair      *Pair     // Asset pair
	Price     uint64    // Last traded price
	Ask       uint64    // Best ask price
	Bid       uint64    // Best bid price
	Volume    uint64    // Trade volume
}

// PotentialPricePoint represents PricePoint that shuold be fetched from Exchange
type PotentialPricePoint struct {
	Pair     *Pair
	Exchange *Exchange
}

// PriceAggregate price aggregation
type PriceAggregate struct {
	Pair   *Pair
	Prices []*PricePoint
	Price  uint64
}

// NewPriceAggregate create new `PriceAggregate`
func NewPriceAggregate(pair *Pair) *PriceAggregate {
	return &PriceAggregate{
		Pair:   pair,
		Prices: []*PricePoint{},
		Price:  0,
	}
}

// Clone clones `PriceAggregate`
func (pa *PriceAggregate) Clone() *PriceAggregate {
	clone := &PriceAggregate{
		Pair:   pa.Pair,
		Prices: make([]*PricePoint, len(pa.Prices)),
		Price:  pa.Price,
	}
	copy(clone.Prices, pa.Prices)
	return clone
}

// ValidateExchange checks if exchange has some error.
// If it's valid no error will be returned, othervise some error.
func ValidateExchange(ex *Exchange) error {
	if ex == nil {
		return fmt.Errorf("exchange is nil")
	}
	if ex.Name == "" {
		return fmt.Errorf("exchange has no name")
	}
	return nil
}

// ValidatePair checks if `Pair` has some errors.
// If it's valid no error will be returned, othervise some error.
func ValidatePair(p *Pair) error {
	if p == nil {
		return fmt.Errorf("pair is nil")
	}
	if p.Base == "" {
		return fmt.Errorf("pair has empty Base")
	}
	if p.Quote == "" {
		return fmt.Errorf("pair has empty Quote")
	}
	return nil
}

// ValidatePotentialPricePoint checks if given `PotentialPricePoint` is valid.
// If it's valid no error will be returned, othervise some error.
func ValidatePotentialPricePoint(pp *PotentialPricePoint) error {
	if pp == nil {
		return fmt.Errorf("given PotentialPricePoint is nil")
	}
	err := ValidateExchange(pp.Exchange)
	if err != nil {
		return fmt.Errorf("given PotentialPricePoint has wrong exchange: %s", err)
	}
	return ValidatePair(pp.Pair)
}

// PriceFromFloat convert price from float value to uint
func PriceFromFloat(f float64) uint64 {
	return uint64(math.Round(f * priceMultiplier))
}

// PriceToFloat convert given `uint64` price to human readable form
func PriceToFloat(price uint64) float64 {
	return float64(price) / priceMultiplier
}
