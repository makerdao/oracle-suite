package gofer

import "fmt"

// Pair
type Pair struct {
	Base  string
	Quote string
}

// Exchange
type Exchange struct {
	Name   string
	Config map[string]string
}

// PotentialPricePoint
type PotentialPricePoint struct {
	Pair     *Pair
	Exchange *Exchange
}

// PricePoint
type PricePoint struct {
	Exchange  *Exchange
	Pair      *Pair
	Timestamp int64
	Price     float64
	Bid       float64
	Ask       float64
	Volume    float64
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
