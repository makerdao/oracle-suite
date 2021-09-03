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

package origins

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/makerdao/oracle-suite/internal/query"
)

const curveURL = "https://api.thegraph.com/subgraphs/name/curvefi/curve"

type curveResponse struct {
	Data struct {
		Pool struct {
			ID            string        `json:"id"`
			A             stringAsInt64 `json:"A"`
			HourlyVolumes []struct {
				Volume stringAsFloat64 `json:"volume"`
			}
			Coins []struct {
				Balance stringAsFloat64 `json:"balance"`
				Token   struct {
					Symbol string `json:"symbol"`
				}
			}
		} `json:"pool"`
	} `json:"data"`
}

type Curve struct {
	WorkerPool query.WorkerPool
}

func (c Curve) Pool() query.WorkerPool {
	return c.WorkerPool
}

func (c *Curve) findPoolForPair(pair Pair) (*curvePool, error) {
	// Because there is no guarantee that other pools would work with current
	// algorithm, it is impossible to use different pools than testes ones.
	if pair.String() == "STETH/ETH" || pair.String() == "ETH/STETH" {
		return &curvePool{
			address:    "0xdc24316b9ae028f1497c275eb9192a3ea0f67022",
			aPrecision: 100,
		}, nil
	}
	return nil, fmt.Errorf("failed to find Curve contract address for pair: %s", pair.String())
}

func (c Curve) PullPrices(pairs []Pair) []FetchResult {
	return callSinglePairOrigin(&c, pairs)
}

func (c *Curve) callOne(pair Pair) (*Price, error) {
	var err error

	pool, err := c.findPoolForPair(pair)
	if err != nil {
		return nil, err
	}
	gql := `
		query($id:String) {
			pool(id: $id) {
				id
				A
				hourlyVolumes(orderBy: timestamp, orderDirection: desc) {
					volume
				}
				coins(orderBy: index, orderDirection: asc) {
					token {
						symbol
					}
					balance
					rate
				}
			}
		}
	`
	body := fmt.Sprintf(
		`{"query":"%s","variables":{"id":"%s"}}`,
		strings.ReplaceAll(strings.ReplaceAll(gql, "\n", " "), "\t", ""),
		pool.address,
	)

	// make query
	req := &query.HTTPRequest{
		URL:    curveURL,
		Method: "POST",
		Body:   bytes.NewBuffer([]byte(body)),
	}
	res := c.Pool().Query(req)
	if res == nil {
		return nil, ErrEmptyOriginResponse
	}
	if res.Error != nil {
		return nil, res.Error
	}

	// parse JSON
	var resp curveResponse
	err = json.Unmarshal(res.Body, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Curve response: %w", err)
	}

	// invalid response
	if resp.Data.Pool.ID != pool.address {
		return nil, ErrMissingResponseForPair
	}

	// find indexes for pairs
	baseIndex := -1
	quoteIndex := -1
	for i, c := range resp.Data.Pool.Coins {
		if strings.ToUpper(c.Token.Symbol) == strings.ToUpper(pair.Base) {
			baseIndex = i
		}
		if strings.ToUpper(c.Token.Symbol) == strings.ToUpper(pair.Quote) {
			quoteIndex = i
		}
	}
	if baseIndex < 0 || quoteIndex < 0 {
		return nil, ErrMissingResponseForPair
	}

	// calculate volume
	volume := float64(0)
	for n, v := range resp.Data.Pool.HourlyVolumes {
		volume += v.Volume.val()
		if n >= 24 {
			break
		}
	}

	// create model and calculate price
	xp := make([]float64, 2)
	for i := 0; i < 2; i++ {
		xp[i] = resp.Data.Pool.Coins[i].Balance.val()
	}
	m := curveModel{
		a:  float64(resp.Data.Pool.A) / pool.aPrecision,
		xp: xp,
	}
	price := m.CalcDY(baseIndex, quoteIndex, 1)

	return &Price{
		Pair:      pair,
		Price:     price,
		Bid:       price,
		Ask:       price,
		Volume24h: volume,
		Timestamp: time.Now(),
	}, nil
}

type curvePool struct {
	address    string  // contract address
	aPrecision float64 // amplification coefficient precision
}

// curveModel calculates values using stableswap invariant.
type curveModel struct {
	a  float64   // amplification coefficient
	xp []float64 // token balances
}

// CalcD calculates D invariant.
func (c curveModel) CalcD() float64 {
	dPrev := float64(0)
	s := c.sum(c.xp)
	d := s
	ann := c.a * float64(c.N())
	for math.Abs(d-dPrev) > 1 {
		dp := d
		for _, x := range c.xp {
			dp = (dp * d) / (float64(c.N()) * x)
		}
		dPrev = d
		d = (ann*s + dp*float64(c.N())) * d / ((ann-1)*d + (float64(c.N())+1)*dp)
	}
	return d
}

// CalcY calculates new balance of j coin when i coin balance become x.
func (c curveModel) CalcY(i, j int, x float64) float64 {
	d := c.CalcD()
	var xx []float64
	for k, _ := range c.xp {
		if k == j {
			continue
		}
		if k == i {
			xx = append(xx, x)
		} else {
			xx = append(xx, c.xp[k])
		}
	}
	ann := c.a * float64(c.N())
	cc := d
	for _, y := range xx {
		cc = cc * d / (y * float64(c.N()))
	}
	cc = cc * d / (float64(c.N()) * ann)
	b := c.sum(xx) + d/ann - d
	yPrev := float64(0)
	y := d
	for math.Abs(y-yPrev) > 1 {
		yPrev = y
		y = (y*y + cc) / (2*y + b)
	}
	return y
}

// CalcDY calculates j coin balance delta when i coin balance changes by dx.
func (c curveModel) CalcDY(i, j int, dx float64) float64 {
	return c.xp[j] - c.CalcY(i, j, c.xp[i]+dx)
}

func (c curveModel) N() int {
	return len(c.xp)
}

func (c curveModel) sum(x []float64) float64 {
	var s float64
	for _, v := range x {
		s += v
	}
	return s
}
