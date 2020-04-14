package reducer

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"makerdao/gofer/model"
)

func TestMedainPriceModel(t *testing.T) {
	rows := []*model.PricePoint{
		// Should be filtered due to outside time window
		{
			Timestamp: -1000,
			Exchange:  "exchange0",
			Base:      "a",
			Quote:     "b",
			Price:     1000,
			Volume:    1,
		},
		// Should be overwritten by entry 3
		{
			Timestamp: 1,
			Exchange:  "exchange1",
			Base:      "a",
			Quote:     "b",
			Price:     1,
			Volume:    1,
		},
		{
			Timestamp: 2,
			Exchange:  "exchange2",
			Base:      "a",
			Quote:     "b",
			Price:     2,
			Volume:    1,
		},
		{
			Timestamp: 3,
			Exchange:  "exchange1",
			Base:      "a",
			Quote:     "b",
			Price:     3,
			Volume:    1,
		},
		// Should not be skipped due to wrong pair
		{
			Timestamp: 4,
			Exchange:  "exchange4",
			Base:      "not",
			Quote:     "reduced",
			Price:     4,
			Volume:    1,
		},
		{
			Timestamp: 5,
			Exchange:  "exchange5",
			Base:      "a",
			Quote:     "b",
			Price:     5,
			Volume:    1,
		},
	}

	reducer := NewMedianReducer(1000)
	priceAggregate := model.NewPriceAggregate("a", "b")

	for _, price := range rows {
		priceAggregate = reducer.Reduce(priceAggregate, price)
	}

	assert.Equal(t, int64(5), priceAggregate.NewestTimestamp, "newest timestamp")
	assert.Equal(t, 3, len(priceAggregate.Prices), "length of aggregate price list")
	assert.Equal(t, uint64(3), priceAggregate.Price, "aggregate price should be median of price points")
}
