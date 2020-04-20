package exchange

import (
	"encoding/json"
	"fmt"
	"makerdao/gofer/model"
	"makerdao/gofer/query"
	"strings"
	"time"
)

// Bitfinex URL
const bitfinexURL = "https://api-pub.bitfinex.com/v2/ticker/t%s"

type bitfinexResponse struct {
	Symbol string `json:"symbol"`
	Price  string `json:"price"`
}

// Bitfinex exchange handler
type Bitfinex struct{}

// Call implementation
func (b *Bitfinex) Call(pool query.WorkerPool, pp *model.PotentialPricePoint) (*model.PricePoint, error) {
	if pool == nil {
		return nil, errNoPoolPassed
	}
	err := model.ValidatePotentialPricePoint(pp)
	if err != nil {
		return nil, err
	}

	pair := strings.ToUpper(pp.Pair.Base + pp.Pair.Quote)
	req := &query.HTTPRequest{
		URL: fmt.Sprintf(bitfinexURL, pair),
	}

	// make query
	res := pool.Query(req)
	if res == nil {
		return nil, errEmptyExchangeResponse
	}
	if res.Error != nil {
		return nil, res.Error
	}
	// parsing JSON
	var resp []float64
	err = json.Unmarshal(res.Body, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to pargse bitfinex response: %s", err)
	}
	if len(resp) < 8 {
		return nil, fmt.Errorf("wrong bitfinex response")
	}
	// building PricePoint
	return &model.PricePoint{
		Exchange:  pp.Exchange,
		Pair:      pp.Pair,
		Price:     model.PriceFromFloat(resp[6]),
		Volume:    model.PriceFromFloat(resp[7]),
		Timestamp: time.Now().Unix(),
	}, nil
}
