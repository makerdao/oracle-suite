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

package aggregator

import (
	"fmt"

	"github.com/makerdao/gofer/model"
)

//  {
//    "b/c": {
//      "method": "median",
//      "sources": [
//        { "exchange": "e-a", "pair": "b/c" },
//        { "exchange": "e-b", "pair": "b/c" },
//      ]
//    },
//    "a/c": {
//      "method": "median",
//      "sources": [
//        { "exchange": "e-a", "pair": "a/c" },
//        { "exchange": "e-a",
//          "pair": "a/b",
//          "op": "multiply",
//          "ref": { "pair": "b/c" }
//        },
//      ]
//    }
//  }

type CacheGetter interface {
	Get(string, Pair) *model.PriceAggregate
}

type cacheID struct {
	pair         Pair
	exchangeName string
}

type PriceCache map[cacheID]*model.PriceAggregate

func (pc PriceCache) Get(exchange string, pair Pair) *model.PriceAggregate {
	pa := pc[cacheID{exchangeName: exchange, pair: pair}]
	if pa == nil {
		return nil
	}
	return pa.Clone()
}

type Pair struct {
	model.Pair
}

type PriceOp int

func (op PriceOp) String() string {
	switch op {
	case MULTIPLY:
		return "*"
	case DIVIDE:
		return "/"
	}
	return "noop"
}

const (
	NOOP PriceOp = iota
	MULTIPLY
	DIVIDE
)

type PriceRef struct {
	ExchangeName string    `json:"exchange,omitempty"`
	Pair         *Pair     `json:"pair,omitempty"`
	Op           PriceOp   `json:"op,omitempty"`
	Ref          *PriceRef `json:"ref,omitempty"`
}

func (pr *PriceRef) String() string {
	var ref string
	if pr.Ref != nil {
		ref = pr.Ref.String()
	}
	if pr.ExchangeName != "" {
		return fmt.Sprintf("{<- (%s %s) %s %s}", pr.ExchangeName, pr.Pair, pr.Op, ref)
	}
	if pr.Op != NOOP {
		return fmt.Sprintf("{-> %s %s %s}", pr.Pair, pr.Op, ref)
	}
	return fmt.Sprintf("{-> %s}", pr.Pair)
}

type PriceModel struct {
	Method  string      `json:"method"`
	Sources []*PriceRef `json:"sources"`
}

type PriceModelMap map[Pair]PriceModel

func (pmm PriceModelMap) getRefSources(ref *PriceRef) []cacheID {
	cis := make(map[cacheID]bool)

	if ref.ExchangeName != "" {
		// PriceRef has an exchange, then return a direct source
		pair := ref.Pair
		ci := cacheID{
			pair:         *pair,
			exchangeName: ref.ExchangeName,
		}
		cis[ci] = true
	} else {
		// PriceRef doesn't have an exchange, then recursivley get price sources
		m, ok := pmm[*ref.Pair]
		if !ok {
			return nil
		}

		for _, pr := range m.Sources {
			for _, ci := range pmm.getRefSources(pr) {
				cis[ci] = true
			}
		}
	}

	if ref.Ref != nil {
		// PriceRef has a reference to another PriceRef
		for _, ci := range pmm.getRefSources(ref.Ref) {
			cis[ci] = true
		}
	}

	var result []cacheID
	for ci := range cis {
		result = append(result, ci)
	}
	return result
}

func (pmm PriceModelMap) GetRefSources(exchangeMap map[string]*model.Exchange, refs ...*PriceRef) []*model.PotentialPricePoint {
	cis := make(map[cacheID]bool)
	for _, ref := range refs {
		for _, ci := range pmm.getRefSources(ref) {
			cis[ci] = true
		}
	}

	var result []*model.PotentialPricePoint

	for ci := range cis {
		exchange := exchangeMap[ci.exchangeName]
		if exchange == nil {
			exchange = &model.Exchange{Name: ci.exchangeName}
		}
		result = append(result, &model.PotentialPricePoint{
			Pair:     ci.pair.Clone(),
			Exchange: exchange,
		})
	}
	return result
}

// GetSources returns PotentialPricePoints
func (pmm PriceModelMap) GetSources(exchangeMap map[string]*model.Exchange) []*model.PotentialPricePoint {
	var refs []*PriceRef
	for pair := range pmm {
		refs = append(refs, &PriceRef{Pair: &pair})
	}
	return pmm.GetRefSources(exchangeMap, refs...)
}

func (pmm PriceModelMap) execOp(pa *model.PriceAggregate, op PriceOp, ref *PriceRef, cache CacheGetter) *model.PriceAggregate {
	if pa == nil {
		return nil
	}
	switch op {
	case MULTIPLY:
		refPa := pmm.ResolveRef(ref, cache)
		if refPa == nil {
			return nil
		}
		return &model.PriceAggregate{
			PricePoint: &model.PricePoint{
				Pair: &model.Pair{
					Base:  pa.Pair.Base,
					Quote: refPa.Pair.Quote,
				},
				Price: pa.Price * refPa.Price,
			},
			Prices:         []*model.PriceAggregate{pa.Clone(), refPa.Clone()},
			PriceModelName: "*",
		}
	case DIVIDE:
		refPa := pmm.ResolveRef(ref, cache)
		if refPa == nil {
			return nil
		}
		return &model.PriceAggregate{
			PricePoint: &model.PricePoint{
				Pair: &model.Pair{
					Base:  pa.Pair.Quote,
					Quote: refPa.Pair.Quote,
				},
				Price: refPa.Price / pa.Price,
			},
			Prices:         []*model.PriceAggregate{refPa.Clone(), pa.Clone()},
			PriceModelName: "/",
		}
	case NOOP:
		return pa.Clone()
	}
	return nil
}

func (pmm PriceModelMap) ResolveRef(ref *PriceRef, cache CacheGetter) *model.PriceAggregate {
	if ref.ExchangeName != "" {
		pair := ref.Pair
		pa := cache.Get(ref.ExchangeName, *pair)
		return pmm.execOp(pa, ref.Op, ref.Ref, cache)
	}

	rootPair := *ref.Pair
	m, ok := pmm[rootPair]
	if !ok {
		return nil
	}

	result := model.PriceAggregate{
		PricePoint: &model.PricePoint{Pair: rootPair.Clone()},
	}
	for _, pr := range m.Sources {
		// Default pair in PriceRef to context pair
		if pr.Pair == nil {
			pr.Pair = &rootPair
		}
		pa := pmm.ResolveRef(pr, cache)
		if pa == nil {
			return nil
		}
		result.Prices = append(result.Prices, pa)
	}

	// TODO: selectable method
	switch m.Method {
	case "":
		fallthrough
	case "median":
		var prices []float64
		for _, pa := range result.Prices {
			prices = append(prices, pa.Price)
		}
		result.Price = median(prices)
		result.PriceModelName = "median"
	default:
		// TODO: replace panic with returning an error
		panic(fmt.Sprintf("price model method '%s' isn't supported", m.Method))
	}

	return &result
}
