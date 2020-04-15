package reducer

import "makerdao/gofer/model"

type Reducer interface {
	// Add a price point to be aggregated
	Ingest(*model.PricePoint)
	// Calculate and return aggregate
	Reduce() *model.PriceAggregate
}
