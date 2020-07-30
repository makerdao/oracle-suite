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
	"log"
	"strings"

	"github.com/makerdao/gofer/model"
)

type CacheGetter interface {
	Get(string, Pair) (*model.PriceAggregate, error)
}

type cacheID struct {
	pair         Pair
	exchangeName string
}

type PriceCache map[cacheID]*model.PriceAggregate

func (pc PriceCache) Get(exchange string, pair Pair) (*model.PriceAggregate, error) {
	ci := cacheID{exchangeName: exchange, pair: pair}
	pa, ok := pc[ci]
	if !ok {
		return nil, fmt.Errorf("key '%s' in price cache not found", ci)
	}
	return pa.Clone(), nil
}

type Pair struct {
	model.Pair
}

type PriceRef struct {
	// TODO: Origin is synonyous with exchange name but needs to map to an exchange
	Origin string `json:"origin,omitempty"`
	Pair   Pair   `json:"pair,omitempty"`
}

func (p *PriceRef) String() string {
	if p.Origin == "." {
		return p.Pair.String()
	}
	return fmt.Sprintf("%s@%s", p.Pair.String(), p.Origin)
}

type PriceRefPath []PriceRef

func (p PriceRefPath) String() string {
	var b strings.Builder
	b.WriteString("[")
	for i, ref := range p {
		if i > 0 {
			b.WriteString(",")
		}
		b.WriteString(ref.String())
	}
	b.WriteString("]")
	return b.String()
}

type PriceModel struct {
	Method     string         `json:"method"`
	MinSources int            `json:"minSourceSuccess"`
	Sources    []PriceRefPath `json:"sources"`
}

func (p *PriceModel) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("(%s:%d)[", p.Method, p.MinSources))
	for _, ref := range p.Sources {
		b.WriteString(ref.String())
	}
	b.WriteString("]")
	return b.String()
}

type PriceModelMethod func([]*model.PriceAggregate) (float64, error)

var PriceModelMethods = map[string]PriceModelMethod{}

type PriceModelMap map[Pair]PriceModel

func (pmm PriceModelMap) getRefSources(pr PriceRef) ([]cacheID, error) {
	cis := make(map[cacheID]bool)

	if pr.Origin == "." {
		// PriceRef is pointing recursivley back to price model map
		m, ok := pmm[pr.Pair]
		if !ok {
			return nil, fmt.Errorf("no pair '%s' in price model map found", pr.Pair.String())
		}

		for _, prp := range m.Sources {
			for _, pr := range prp {
				cs, err := pmm.getRefSources(pr)
				if err != nil {
					return nil, err
				}
				for _, ci := range cs {
					cis[ci] = true
				}
			}
		}
	} else {
		// PriceRef has an origin, then return a direct source
		pair := pr.Pair
		ci := cacheID{
			pair:         pair,
			exchangeName: pr.Origin,
		}
		cis[ci] = true
	}

	var result []cacheID
	for ci := range cis {
		result = append(result, ci)
	}
	return result, nil
}

func (pmm PriceModelMap) GetRefSources(exchangeMap map[string]*model.Exchange, refs ...PriceRef) ([]*model.PotentialPricePoint, error) {
	cis := make(map[cacheID]bool)
	for _, ref := range refs {
		cs, err := pmm.getRefSources(ref)
		if err != nil {
			return nil, err
		}
		for _, ci := range cs {
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
	return result, nil
}

// GetSources returns PotentialPricePoints
func (pmm PriceModelMap) GetSources(exchangeMap map[string]*model.Exchange) ([]*model.PotentialPricePoint, error) {
	var refs []PriceRef
	for pair := range pmm {
		refs = append(refs, PriceRef{Origin: ".", Pair: pair})
	}
	return pmm.GetRefSources(exchangeMap, refs...)
}

// Calculate the final model.trade price of an ordered list of prices
func resolvePath(pas []*model.PriceAggregate) (*model.PriceAggregate, error) {
	if len(pas) == 1 {
		return pas[0], nil
	}

	var pair *model.Pair
	var price float64

	for _, pa := range pas {
		if pair == nil {
			pair = pa.Pair.Clone()
			price = pa.Price
		} else if pair.Quote == pa.Pair.Quote {
			price = pa.Price / price
			pair.Quote = pa.Pair.Base
		} else if pair.Quote == pa.Pair.Base {
			price *= pa.Price
			pair.Quote = pa.Pair.Quote
		} else {
			return nil, fmt.Errorf("can't convert between %s and %s", pair, pa.Pair)
		}
	}

	return model.NewPriceAggregate(
		"trade",
		&model.PricePoint{
			Pair:  pair,
			Price: price,
		},
		pas...,
	), nil
}

func (pmm PriceModelMap) resolvePath(cache CacheGetter, prp PriceRefPath) (*model.PriceAggregate, error) {
	var pas []*model.PriceAggregate
	for _, pr := range prp {
		pa, err := pmm.ResolveRef(cache, pr)
		if err != nil {
			return nil, err
		}
		pas = append(pas, pa)
	}
	pa, err := resolvePath(pas)
	if err != nil {
		return nil, err
	}
	return pa, nil
}

func (pmm PriceModelMap) ResolveRef(cache CacheGetter, pr PriceRef) (*model.PriceAggregate, error) {
	if pr.Origin != "." {
		return cache.Get(pr.Origin, pr.Pair)
	}

	m, ok := pmm[pr.Pair]
	if !ok {
		return nil, fmt.Errorf("no pair '%s' in price model map found", pr.Pair.String())
	}

	result := model.PriceAggregate{
		PricePoint: &model.PricePoint{Pair: pr.Pair.Clone()},
	}
	for _, prp := range m.Sources {
		pa, err := pmm.resolvePath(cache, prp)
		if err != nil {
			// TODO: log sources that couldn't be resolved?
			log.Print(err)
			continue
		}
		if !pr.Pair.Equal(pa.Pair) {
			// TODO: log error when indirect pair of resolved path doesn't match price
			// model key
			log.Printf(
				"failed to resolve source %s, %s != %s\n\t%s\n",
				pr, pr.Pair.String(), pa.Pair.String(), pa)
			continue
		}
		// Add price aggregate if resolved successfully
		result.Prices = append(result.Prices, pa)
	}

	successes := len(result.Prices)
	if successes == 0 || successes < m.MinSources {
		return nil, fmt.Errorf("minimum amount of sources not met for '%s' in price model", pr.Pair.String())
	}

	method, ok := PriceModelMethods[m.Method]
	if !ok {
		return nil, fmt.Errorf("price model method '%s' isn't supported", m.Method)
	}

	var err error
	result.Price, err = method(result.Prices)
	if err != nil {
		return nil, err
	}

	result.PriceModelName = m.Method

	return &result, nil
}

func medianPriceModelMethod(pas []*model.PriceAggregate) (float64, error) {
	var prices []float64
	for _, p := range pas {
		prices = append(prices, p.Price)
	}
	return median(prices), nil
}

func init() {
	PriceModelMethods[""] = medianPriceModelMethod
	PriceModelMethods["median"] = medianPriceModelMethod
}
