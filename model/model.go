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
	"strings"
)

// Pair represents currency pair
type Pair struct {
	Base  string
	Quote string
}

// Exchange represent exchange details
type Exchange struct {
	Name   string            `json:"name"`
	Config map[string]string `json:"parameters"`
}

// PricePoint given price point
type PricePoint struct {
	Timestamp int64     // Unix time
	Exchange  *Exchange // Exchange id
	Pair      *Pair     // Asset pair
	Price     float64    // Last traded price
	Ask       float64    // Best ask price
	Bid       float64    // Best bid price
	Volume    float64    // Trade volume
}

// PotentialPricePoint represents PricePoint that shuold be fetched from Exchange
type PotentialPricePoint struct {
	Pair     *Pair
	Exchange *Exchange
}

// String returns a string representation of `PotentialPricePoint` e.g. source[exchange](BTC/USD)
func (ppp *PotentialPricePoint) String() string {
	var pair string
	var exchange string
	if ppp.Exchange != nil {
		exchange = ppp.Exchange.Name
	}
	if ppp.Pair != nil {
		pair = ppp.Pair.String()
	}
	return fmt.Sprintf("source[%s](%s)", exchange, pair)
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

// PricePathMap represents a set of PricePath indexed by their target pair, each
// pair can have multiples paths
type PricePathMap map[Pair][]*PricePath

// Target returns a Pair with base of first and quote of last pair in path
func (ppath PricePath) Target() *Pair {
	var pair *Pair
	for _, p := range ppath {
		if pair == nil {
			pair = p.Clone()
		} else if pair.Base == p.Base {
			pair.Base = pair.Quote
			pair.Quote = p.Quote
		} else if pair.Quote == p.Base {
			pair.Quote = p.Quote
		} else {
			return nil
		}
	}

	return pair
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
	return &Pair{Base: strings.ToUpper(base), Quote: strings.ToUpper(quote)}
}

// Equal check if `Pair` is equal to given one
func (p *Pair) Equal(pair *Pair) bool {
	return (p.Base == pair.Base && p.Quote == pair.Quote)
}

// String returns a string representation of `Pair` e.g. BTC/USD
func (p *Pair) String() string {
	return fmt.Sprintf("%s/%s", p.Base, p.Quote)
}

// Create a copy of Pair
func (p *Pair) Clone() *Pair {
	return &Pair{Base: p.Base, Quote: p.Quote}
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
		return fmt.Errorf("given PotentialPricePoint has wrong exchange: %w", err)
	}
	return ValidatePair(pp.Pair)
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

// NewPricePathMap creates a new instance of `PricePaths`
func NewPricePathMap(ppaths []*PricePath) PricePathMap {
	ppaths_ := make(PricePathMap)
	for _, ppath := range ppaths {
		target := *ppath.Target()
		ppaths_[target] = append(ppaths_[target], ppath)
	}
	return ppaths_
}

// String returns PricePaths string representation
func (ppaths PricePathMap) String() string {
	var str strings.Builder
	for pair, ppaths_ := range ppaths {
		str.WriteString(pair.String())
		str.WriteString(" =")
		for _, path := range ppaths_ {
			str.WriteString(" (")
			str.WriteString(path.String())
			str.WriteString(")")
		}
		str.WriteString("\n")
	}
	return str.String()
}

// ValidatePricePath checks if price path has at least one pair
func ValidatePricePath(ppath *PricePath) error {
	if ppath == nil {
		return fmt.Errorf("price path is nil")
	}

	for _, p := range *ppath {
		if ValidatePair(p) != nil {
			return fmt.Errorf("pair in path is invalid")
		}
	}

	if ppath.Target() == nil {
		return fmt.Errorf("price path is not valid")
	}

	return nil
}

// ValidatePricePathMap checks if price paths all have matching target pairs and
// paths are valid
func ValidatePricePathMap(ppaths PricePathMap) error {
	if ppaths == nil {
		return fmt.Errorf("price paths is nil")
	}

	for pair, ppaths_ := range ppaths {
		if err := ValidatePair(&pair); err != nil {
			return fmt.Errorf("a target pair is invalid: %w", err)
		}

		if ppaths_ == nil {
			return fmt.Errorf("no paths for pair %s", pair)
		}

		for _, ppath := range ppaths_ {
			if err := ValidatePricePath(ppath); err != nil {
				return fmt.Errorf("a path for pair %s is invalid: %w", pair, err)
			}

			target := ppath.Target()
			if err := ValidatePair(target); err != nil {
				return fmt.Errorf("a path for pair %s has invalid target pair: %w", pair, err)
			}

			if !pair.Equal(target) {
				return fmt.Errorf("a path for pair %s has mismatching target pair %s", pair, target)
			}
		}
	}
	return nil
}

// String returns PricePath string representation
func (pa *PriceAggregate) String() string {
	var str strings.Builder
	str.WriteString(fmt.Sprintf("%f", pa.Price))
	str.WriteString("$")
	str.WriteString(pa.Pair.String())
	str.WriteString("<=")
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
	str.WriteString(")")
	return str.String()
}
