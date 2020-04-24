package gofer

import (
	"github.com/stretchr/testify/assert"
	"makerdao/gofer/model"
	"makerdao/gofer/query"
	"testing"
)

func TestPriceCollectorFailsWithoutWorkingPool(t *testing.T) {
	p := &model.Pair{
		Base:  "BTC",
		Quote: "ETH",
	}
	pp := &model.PotentialPricePoint{
		Exchange: &model.Exchange{
			Name: "binance",
		},
		Pair: p,
	}

	pc := &PriceCollector{}
	_, err := pc.CollectPricePoint(pp)
	assert.Error(t, err)

	// Check not started
	pc = &PriceCollector{
		wp: &query.HTTPWorkerPool{},
	}
	_, err = pc.CollectPricePoint(pp)
	assert.Error(t, err)
}
