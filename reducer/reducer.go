package reducer

import "makerdao/gofer/model"

type Reducer interface {
	Reduce(*model.PriceAggregate, *model.PricePoint) *model.PriceAggregate
}
