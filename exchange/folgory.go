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

// Folgory URL
const folgoryURL = "https://www.folgory.com/api/v3/ticker/price?symbol=%s"

type folgoryResponse struct {
	Symbol string `json:"symbol"`
	Price  string `json:"last"`
	Volume string `json:"volume"`
}

// Folgory exchange handler
type Folgory struct{}

// Call implementation
func (b *Folgory) Call(pool query.WorkerPool, pp *model.PotentialPricePoint) (*model.PricePoint, error) {
	if pool == nil {
		return nil, errNoPoolPassed
	}
	err := model.ValidatePotentialPricePoint(pp)
	if err != nil {
		return nil, err
	}

	pair := fmt.Sprintf("%s/%s", strings.ToUpper(pp.Pair.Base), strings.ToUpper(pp.Pair.Quote))
	req := &query.HTTPRequest{
		URL: fmt.Sprintf(folgoryURL, pair),
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
	var resp []folgoryResponse
	err = json.Unmarshal(res.Body, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to pargse folgory response: %s", err)
	}

	var data *folgoryResponse
	for _, symbol := range resp {
		if symbol.Symbol == pair {
			data = &symbol
			break
		}
	}
	if data == nil {
		return nil, fmt.Errorf("wrong response from folgory. no %s pair exist", pair)
	}
	// Parsing price from string
	price, err := strconv.ParseFloat(data.Price, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse price from folgory exchange %v", data)
	}
	// Parsing volume from string
	volume, err := strconv.ParseFloat(data.Volume, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse volume from folgory exchange %v", data)
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
