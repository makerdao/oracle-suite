package gofer

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
