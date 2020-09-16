//  Copyright (C) 2020 Maker Ecosystem Growth Holdings, INC.
//
//  This program is free software: you can redistribute it and/or modify
//  it under the terms of the GNU Affero General Public License as
//  published by the Free Software Foundation, either version 3 of the
//  License, or (at your option) any later version.
//
//  This program is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU Affero General Public License for more details.
//
//  You should have received a copy of the GNU Affero General Public License
//  along with this program.  If not, see <http://www.gnu.org/licenses/>.

package exchange

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/makerdao/gofer/internal/query"
	"github.com/makerdao/gofer/pkg/model"
)

//nolint:lll
const cryptoCompareMultiURL = "https://min-api.cryptocompare.com/data/pricemultifull?fsyms=%s&tsyms=%s&tryConversion=false&extraParams=gofer&relaxedValidation=true"

type cryptoCompareMultiResponse struct {
	Raw map[string]map[string]struct {
		Base  string  `json:"FROMSYMBOL"`
		Query string  `json:"TOSYMBOL"`
		Price float64 `json:"PRICE"`
		TS    int64   `json:"LASTUPDATE"`
		Vol24 float64 `json:"VOLUME24HOUR"`
	} `json:"RAW"`
}

type CryptoCompare struct {
	Pool query.WorkerPool
}

func (c *CryptoCompare) Call(ppps []*model.PotentialPricePoint) []CallResult {
	req := c.makeRequest(ppps)
	res := c.Pool.Query(req)
	if errorResponses := validateResponse(ppps, res); len(errorResponses) > 0 {
		return errorResponses
	}
	return c.parseResponse(ppps, res)
}

func (c *CryptoCompare) makeRequest(ppps []*model.PotentialPricePoint) *query.HTTPRequest {
	var bList, qList []string

	for _, ppp := range ppps {
		bList = append(bList, ppp.Pair.Base)
	}
	for _, ppp := range ppps {
		qList = append(qList, ppp.Pair.Quote)
	}

	req := &query.HTTPRequest{
		URL: fmt.Sprintf(cryptoCompareMultiURL, strings.Join(bList, ","), strings.Join(qList, ",")),
	}
	return req
}

func (c *CryptoCompare) parseResponse(ppps []*model.PotentialPricePoint, res *query.HTTPResponse) []CallResult {
	results := make([]CallResult, 0)

	var resp cryptoCompareMultiResponse
	err := json.Unmarshal(res.Body, &resp)
	if err != nil {
		for _, ppp := range ppps {
			results = append(results, newCallResult(
				ppp,
				nil,
				fmt.Errorf("failed to parse CryptoCompare response: %w", err),
			))
		}
		return results
	}

	for _, ppp := range ppps {
		if bObj, is := resp.Raw[ppp.Pair.Base]; !is {
			results = append(results, newCallResult(
				ppp,
				nil,
				fmt.Errorf("no response for %s base from %s", ppp.Pair.Base, ppp.Exchange.Name),
			))
		} else {
			if qObj, is := bObj[ppp.Pair.Quote]; !is {
				results = append(results, newCallResult(
					ppp,
					nil,
					fmt.Errorf("no response for %s quote from %s", ppp.Pair.Quote, ppp.Exchange.Name),
				))
			} else {
				pp := &model.PricePoint{
					Timestamp: qObj.TS,
					Exchange:  ppp.Exchange,
					Pair:      model.NewPair(qObj.Base, qObj.Query),
					Price:     qObj.Price,
					Volume:    qObj.Vol24,
				}
				results = append(results, newCallResult(ppp, pp, nil))
			}
		}
	}
	return results
}
