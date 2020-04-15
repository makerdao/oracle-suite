package model

type Pair struct {
	Base  string
	Quote string
}

func (p *Pair) Equal(pair *Pair) bool {
	return p.Base == pair.Base && p.Quote == p.Quote
}

type PricePoint struct {
	Timestamp int64
	Exchange  string
	Pair      *Pair
	Price     uint64
	Volume    uint64
}

type PriceAggregate struct {
	Pair   *Pair
	Prices []*PricePoint
	Price  uint64
}

func NewPriceAggregate(pair *Pair) *PriceAggregate {
	return &PriceAggregate{
		Pair:   pair,
		Prices: []*PricePoint{},
		Price:  0,
	}
}

func (pa *PriceAggregate) Clone() *PriceAggregate {
	clone := &PriceAggregate{
		Pair:   pa.Pair,
		Prices: make([]*PricePoint, len(pa.Prices)),
		Price:  pa.Price,
	}
	copy(clone.Prices, pa.Prices)
	return clone
}
