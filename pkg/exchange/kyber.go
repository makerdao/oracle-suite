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
	"time"

	"github.com/makerdao/gofer/internal/pkg/query"
	"github.com/makerdao/gofer/pkg/model"
)

const kyberURL = "https://api.kyber.network/buy_rate?id=%s&qty=%g"

type kyberDataResponse struct {
	Src    string    `json:"src_id"`
	Dst    string    `json:"dst_id"`
	SrcQty []float64 `json:"src_qty"`
	DstQty []float64 `json:"dst_qty"`
}

type kyberResponse struct {
	Error          bool                 `json:"error"`
	Reason         string               `json:"reason"`
	AdditionalData string               `json:"additional_data"`
	Result         []*kyberDataResponse `json:"data"`
}

type Kyber struct {
	Pool query.WorkerPool
}

func (k *Kyber) localPairName(pair *model.Pair) string {
	var addrList = map[model.Pair]string{
		{Base: "DGX", Quote: "ETH"}:  "0x4f3afec4e5a3f2a6a1a411def7d7dfe50ee057bf",
		{Base: "KNC", Quote: "ETH"}:  "0xdd974d5c2e2928dea5f71b9825b8b646686bd200",
		{Base: "LEND", Quote: "ETH"}: "0x80fB784B7eD66730e8b1DBd9820aFD29931aab03",
		{Base: "MKR", Quote: "ETH"}:  "0x9f8f72aa9304c8b593d555f12ef6589cc3a579a2",
		{Base: "WBTC", Quote: "ETH"}: "0x2260fac5e5542a773aa44fbcfedf7c193bc2c599",
	}
	return addrList[*pair]
}

const refQty = 2.5

func (k *Kyber) getURL(pp *model.PotentialPricePoint) string {
	return fmt.Sprintf(kyberURL, k.localPairName(pp.Pair), refQty)
}

func (k *Kyber) Call(pp *model.PotentialPricePoint) (*model.PricePoint, error) {
	err := model.ValidatePotentialPricePoint(pp)
	if err != nil {
		return nil, err
	}

	req := &query.HTTPRequest{
		URL: k.getURL(pp),
	}

	// make query
	res := k.Pool.Query(req)
	if res == nil {
		return nil, errEmptyExchangeResponse
	}
	if res.Error != nil {
		return nil, res.Error
	}
	// parsing JSON
	var resp kyberResponse
	err = json.Unmarshal(res.Body, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse kyber response: %w", err)
	}
	if resp.Error {
		return nil, fmt.Errorf("kyber API error: %s (%s)", resp.Reason, resp.AdditionalData)
	}
	if len(resp.Result) == 0 {
		return nil, fmt.Errorf("wrong kyber exchange response. No resulting data %+v", resp)
	}
	result := resp.Result[0]

	if len(result.SrcQty) == 0 || len(result.DstQty) == 0 {
		return nil, fmt.Errorf("wrong kyber exchange response. No resulting pair %s data %+v", pp.Pair.String(), result)
	}

	if result.SrcQty[0] <= 0 {
		return nil, fmt.Errorf("failed to parse price from kyber exchange (needs to be gtreater than 0) %s", res.Body)
	}

	if result.DstQty[0] != refQty {
		return nil, fmt.Errorf("failed to parse volume from kyber exchange (it needs to be %f) %s", refQty, res.Body)
	}

	if result.Src != "0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee" {
		return nil, fmt.Errorf("failed to parse price from kyber exchange (src needs to be 0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee) %s", res.Body)
	}

	if result.Dst != k.localPairName(pp.Pair) {
		return nil, fmt.Errorf("failed to parse volume from kyber exchange (it needs to be %f) %s", refQty, res.Body)
	}

	// building PricePoint
	return &model.PricePoint{
		Exchange:  pp.Exchange,
		Pair:      pp.Pair,
		Price:     result.SrcQty[0] / result.DstQty[0],
		Volume:    result.DstQty[0],
		Timestamp: time.Now().Unix(),
	}, nil
}
