package exchange

import (
	"encoding/json"
	"fmt"
	"makerdao/gofer/model"
	"makerdao/gofer/query"
	"strconv"
	"strings"
	"time"
)

// Huobi URL
const huobiURL = "https://api.huobi.pro/market/detail/merged?symbol=%s"

type huobiResponse struct {
	Status string `json:"status"`
	Volume string `json:"vol"`
	Tick   struct {
		Bid []string
	}
}

// Huobi exchange handler
type Huobi struct{}

// Call implementation
func (b *Huobi) Call(pool query.WorkerPool, pp *model.PotentialPricePoint) (*model.PricePoint, error) {
	if pool == nil {
		return nil, errNoPoolPassed
	}
	err := model.ValidatePotentialPricePoint(pp)
	if err != nil {
		return nil, err
	}

	pair := strings.ToLower(pp.Pair.Base + pp.Pair.Quote)
	req := &query.HTTPRequest{
		URL: fmt.Sprintf(huobiURL, pair),
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
	var resp huobiResponse
	err = json.Unmarshal(res.Body, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to pargse huobi response: %s", err)
	}
	if resp.Status == "error" {
		return nil, fmt.Errorf("wrong response from huobi exchange %s", res.Body)
	}
	// Parsing price from string
	price, err := strconv.ParseFloat(resp.Tick.Bid[0], 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse price from huobi exchange %s", res.Body)
	}
	// Parsing volume from string
	volume, err := strconv.ParseFloat(resp.Volume, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse volume from huobi exchange %s", res.Body)
	}
	// building PricePoint
	return &model.PricePoint{
		Exchange:  pp.Exchange,
		Pair:      pp.Pair,
		Price:     model.PriceFromFloat(price),
		Volume:    model.PriceFromFloat(volume),
		Timestamp: time.Now().Unix(),
	}, nil
}
