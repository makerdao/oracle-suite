package reducer

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"makerdao/gofer/model"
)

func TestOddPriceCount(t *testing.T) {
	rows := []*model.PricePoint{
		// Should be filtered due to outside time window
		{-1000, "exchange0", &model.Pair{"a", "b"}, 1000, 1},
		// Should be overwritten by entry 3 due to same exchange but older
		{1, "exchange1", &model.Pair{"a", "b"}, 2000, 1},
		{2, "exchange2", &model.Pair{"a", "b"}, 20, 1},
		{3, "exchange1", &model.Pair{"a", "b"}, 3, 1},
		// Should be skipped due to non-matching pair
		{4, "exchange4", &model.Pair{"n", "o"}, 4, 1},
		{5, "exchange5", &model.Pair{"a", "b"}, 5, 1},
	}

	for i := 0; i < 100; i++ {
		reducer := NewMedianReducer(&model.Pair{"a", "b"}, 1000)
		priceAggregate := RandomReduce(reducer, rows)
		assert.Equal(t, 3, len(priceAggregate.Prices), "length of aggregate price list")
		assert.Equal(t, uint64(5), priceAggregate.Price, "aggregate price should be median of price points")
	}
}

func TestEvenPriceCount(t *testing.T) {
	rows := []*model.PricePoint{
		{1, "exchange1", &model.Pair{"a", "b"}, 7, 1},
		{2, "exchange2", &model.Pair{"a", "b"}, 2, 1},
		{3, "exchange3", &model.Pair{"a", "b"}, 10, 1},
		{4, "exchange4", &model.Pair{"a", "b"}, 5, 1},
	}

	for i := 0; i < 100; i++ {
		reducer := NewMedianReducer(&model.Pair{"a", "b"}, 1000)
		priceAggregate := RandomReduce(reducer, rows)
		assert.Equal(t, 4, len(priceAggregate.Prices), "length of aggregate price list")
		assert.Equal(t, uint64(6), priceAggregate.Price, "aggregate price should be median of price points")
	}
}
