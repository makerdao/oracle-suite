package exchange

import (
	"encoding/json"
	"fmt"
	"makerdao/gofer/model"
	"makerdao/gofer/query"
	"strings"
	"time"
)

// Upbit URL
const upbitURL = "https://api.upbit.com/v1/ticker?markets=%s"

type upbitResponse struct {
	Price  float64 `json:"trade_price"`
	Volume float64 `json:"acc_trade_volume"`
}

// Upbit exchange handler
type Upbit struct{}

// Call implementation
func (b *Upbit) Call(pool query.WorkerPool, pp *model.PotentialPricePoint) (*model.PricePoint, error) {
	if pool == nil {
		return nil, errNoPoolPassed
	}
	err := model.ValidatePotentialPricePoint(pp)
	if err != nil {
		return nil, err
	}

	pair := fmt.Sprintf("%s-%s", strings.ToUpper(pp.Pair.Quote), strings.ToUpper(pp.Pair.Base))
	req := &query.HTTPRequest{
		URL: fmt.Sprintf(upbitURL, pair),
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
	var resp []upbitResponse
	err = json.Unmarshal(res.Body, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to pargse upbit response: %s", err)
	}
	if len(resp) < 1 {
		return nil, fmt.Errorf("wrong upbit response: %s", res.Body)
	}
	data := resp[0]
	// building PricePoint
	return &model.PricePoint{
		Exchange:  pp.Exchange,
		Pair:      pp.Pair,
		Price:     model.PriceFromFloat(data.Price),
		Volume:    model.PriceFromFloat(data.Volume),
		Timestamp: time.Now().Unix(),
	}, nil
}
