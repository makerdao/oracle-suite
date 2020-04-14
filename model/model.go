package model

type PricePoint struct {
	Timestamp int64
	Exchange  string
	Base      string
	Quote     string
	Price     uint64
	Volume    uint64
}

type PriceAggregate struct {
	Base            string
	Quote           string
	Prices          []*PricePoint
	Price           uint64
	NewestTimestamp int64
}

func NewPriceAggregate(base string, quote string) *PriceAggregate {
	return &PriceAggregate{
		Base:            base,
		Quote:           quote,
		Prices:          []*PricePoint{},
		Price:           0,
		NewestTimestamp: 0,
	}
}
