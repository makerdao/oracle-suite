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

// Gateio URL
const gateioURL = "https://fx-api.gateio.ws/api/v4/futures/tickers?contract=%s"

type gateioResponse struct {
	Volume string `json:"volume_24h_base"`
	Price  string `json:"last"`
}

// Gateio exchange handler
type Gateio struct{}

// Call implementation
func (b *Gateio) Call(pool query.WorkerPool, pp *model.PotentialPricePoint) (*model.PricePoint, error) {
	if pool == nil {
		return nil, errNoPoolPassed
	}
	err := model.ValidatePotentialPricePoint(pp)
	if err != nil {
		return nil, err
	}

	pair := fmt.Sprintf("%s_%s", strings.ToUpper(pp.Pair.Base), strings.ToUpper(pp.Pair.Quote))
	req := &query.HTTPRequest{
		URL: fmt.Sprintf(gateioURL, pair),
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
	var resp []gateioResponse
	err = json.Unmarshal(res.Body, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to pargse gateio response: %s", err)
	}
	if len(resp) < 1 {
		return nil, fmt.Errorf("wrong gateio response %s", res.Body)
	}
	// Parsing price from string
	price, err := strconv.ParseFloat(resp[0].Price, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse price from gateio exchange %s", res.Body)
	}
	// Parsing price from string
	volume, err := strconv.ParseFloat(resp[0].Volume, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse volume from gateio exchange %s", res.Body)
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
