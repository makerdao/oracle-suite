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
	"strings"
)

const priceMultiplier = 10000000

// Pair represents currency pair
type Pair struct {
	Base  string
	Quote string
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
	*PricePoint
	Prices         []*PriceAggregate
	PriceModelName string
}

// PricePath represents a continuous chain of asset pairs that can be traded in
// sequence
type PricePath []*Pair

// PricePath represents a way to convert from a base asset to a quote asset
// through direct trading pairs
type PricePaths struct {
	Target *Pair
	Paths  []PricePath
}

// Target returns a Pair with base of first and quote of last pair in path
func (ppath PricePath) Target() *Pair {
	pathLen := len(ppath)
	if pathLen == 0 {
		return nil
	}
	return NewPair(ppath[0].Base, ppath[pathLen-1].Quote)
}

// NewPriceAggregate create new `PriceAggregate`
func NewPriceAggregate(name string, price *PricePoint, prices ...*PriceAggregate) *PriceAggregate {
	return &PriceAggregate{
		PricePoint:     price,
		PriceModelName: name,
		Prices:         prices,
	}
}

// Clone clones `PriceAggregate`
func (pa *PriceAggregate) Clone() *PriceAggregate {
	clone := &PriceAggregate{
		PriceModelName: pa.PriceModelName,
		PricePoint:     &PricePoint{Pair: pa.Pair, Price: pa.Price},
		Prices:         make([]*PriceAggregate, len(pa.Prices)),
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

// NewPair creates a new instance of `Pair`
func NewPair(base string, quote string) *Pair {
	return &Pair{Base: base, Quote: quote}
}

// Equal check if `Pair` is equal to given one
func (p *Pair) Equal(pair *Pair) bool {
	return (p.Base == pair.Base && p.Quote == pair.Quote)
}

// String returns a string representation of `Pair` e.g. BTC/USD
func (p *Pair) String() string {
	return fmt.Sprintf("%s/%s", p.Base, p.Quote)
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
	if p.Base == p.Quote {
		return fmt.Errorf("pair has same Base and Quote")
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

// NewPricePaths creates a new instance of `PricePaths`
func NewPricePaths(target *Pair, pairs ...PricePath) *PricePaths {
	return &PricePaths{
		Target: target,
		Paths:  pairs,
	}
}

// String returns PricePaths string representation
func (ppaths *PricePaths) String() string {
	var str strings.Builder
	str.WriteString(ppaths.Target.String())
	str.WriteString(" =")
	for _, path := range ppaths.Paths {
		str.WriteString(" (")
		str.WriteString(path.String())
		str.WriteString(")")
	}
	return str.String()
}

// String returns PricePath string representation
func (ppath PricePath) String() string {
	var str strings.Builder
	str.WriteString(ppath[0].Base)
	for _, pair := range ppath {
		str.WriteString(" -> ")
		str.WriteString(pair.Quote)
	}
	return str.String()
}

// ValidatePricePath checks if price path has at least one pair
func ValidatePricePath(ppath PricePath) error {
	if ppath == nil {
		return fmt.Errorf("price path is nil")
	}
	if len(ppath) == 0 {
		return fmt.Errorf("price path has no pairs")
	}
	var prev *Pair
	for _, p := range ppath {
		if ValidatePair(p) != nil {
			return fmt.Errorf("pair in path is invalid")
		}
		if prev != nil {
			if prev.Quote != p.Base {
				return fmt.Errorf("price path sequence is invalid")
			}
		}
		prev = p
	}
	return nil
}

// ValidatePricePaths checks if price paths all have same target pair
func ValidatePricePaths(ppaths *PricePaths) error {
	if ppaths == nil {
		return fmt.Errorf("price paths is nil")
	}
	if ValidatePair(ppaths.Target) != nil {
		return fmt.Errorf("price paths target pair is invalid")
	}
	if len(ppaths.Paths) == 0 {
		return fmt.Errorf("price paths has no paths")
	}
	for _, ppath := range ppaths.Paths {
		if ValidatePricePath(ppath) != nil {
			return fmt.Errorf("one path in price paths was invalid")
		}
		if !ppaths.Target.Equal(ppath.Target()) {
			return fmt.Errorf("one path in price paths has non matching target")
		}
	}
	return nil
}

// String returns PricePath string representation
func (pa *PriceAggregate) String() string {
	var str strings.Builder
	str.WriteString(pa.PriceModelName)
	str.WriteString("(")
	count := len(pa.Prices)
	if count > 0 {
		str.WriteString(" ")
		for i, pa_ := range pa.Prices {
			str.WriteString(pa_.String())
			if i < count-1 {
				str.WriteString(",")
			}
			str.WriteString(" ")
		}
	}
	str.WriteString(")=>")
	str.WriteString(pa.Pair.String())
	str.WriteString("$")
	str.WriteString(fmt.Sprintf("%d", pa.Price))
	return str.String()
}

// PriceFromFloat convert price from float value to uint
func PriceFromFloat(f float64) uint64 {
	return uint64(math.Round(f * priceMultiplier))
}

// PriceToFloat convert given `uint64` price to human readable form
func PriceToFloat(price uint64) float64 {
	return float64(price) / priceMultiplier
}
